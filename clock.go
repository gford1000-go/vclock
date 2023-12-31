package vclock

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"

	"github.com/gford1000-go/chant"
	"github.com/gford1000-go/syncmap"
)

// Clock is the underlying type of the vector clock
type Clock map[string]uint64

type AllowedReq interface {
	Clock | *respComp | *reqFullHistory | *reqGet | *reqHistory | *reqLastUpdate | *reqPrune | *reqSnap | *reqSnapShortenedIdentifiers | *SetInfo | *reqTick
}

type AllowedResp interface {
	*respClock | *respErr | bool | *respGetter | *respGetterWithStatus | *respHistory | *respHistoryAll
}

// attemptSendChanWithResp will stop the panic and return recoverErr, should the chan be closed
func attemptSendChanWithResp[T AllowedReq, U AllowedResp](c *chant.Channel[any], t T, r *chant.Channel[any], recoverErr error) (u U, err error) {
	handleChanErr := func(e error) (U, error) {
		var u U
		if !errors.Is(e, chant.ErrChannelClosed) {
			return u, e
		}
		return u, recoverErr
	}

	if err := c.Send(t); err != nil {
		return handleChanErr(err)
	}
	if resp, err := r.Recv(); err != nil {
		return handleChanErr(err)
	} else {
		return resp.(U), nil
	}
}

// attemptSendChan is syntax sugar to simply the call when only an error would be returned
func attemptSendChan[T AllowedReq](c *chant.Channel[any], t T, r *chant.Channel[any], recoverErr error) error {
	resp, err := attemptSendChanWithResp[T, *respErr](c, t, r, recoverErr)
	if err != nil {
		return err
	}
	return resp.err
}

type reqFullHistory struct {
}

type reqGet struct {
	id string
}

type reqHistory struct {
}

type reqLastUpdate struct {
}

type reqPrune struct {
}

type reqSnap struct {
}

type reqSnapShortenedIdentifiers struct {
}

type reqTick struct {
	id string
}

type respClock struct {
	c Clock
	e error
}

type respErr struct {
	err error
}

type respGetter struct {
	id string
	v  uint64
	e  error
}

type respGetterWithStatus struct {
	respGetter
	b bool
}

type respHistory struct {
	h []Clock
	e error
}

type respHistoryAll struct {
	h []*HistoryItem
	e error
}

var errClockIdMustNotBeEmptyString = errors.New("clock identifier must not be empty string")
var errAttemptToSetExistingId = errors.New("clock identifier cannot be reset once initialised")
var errAttemptToTickUnknownId = errors.New("attempted to tick unknown clock identifier")
var errClosedVClock = errors.New("attempt to interact with closed clock")
var errClockMustNotBeNil = errors.New("attempt to merge a nil clock")
var errUnknownReqType = errors.New("received unknown request struct")

// VClock is an instance of a vector clock that can suppport
// concurrent use across multiple goroutines
type VClock struct {
	req       *chant.Channel[any]
	resp      *chant.Channel[any]
	shortener string
	ctx       context.Context
	cancel    context.CancelFunc
}

// New returns a VClock that is initialised with the specified Clock details,
// and which will not maintain any history.  The specified shortener
// (which may be empty string) reduces the memory footprint of the vector
// clock if the identifiers are large strings.
func New(context context.Context, init Clock, shortenerName string) (*VClock, error) {
	return newClock(context, init, false, shortenerName, true)
}

// NewWithHistory returns a VClock that is initialised with the specified Clock details,
// and which will maintain a full history of all updates.  The specified shortener
// (which may be empty string) reduces the memory footprint of the vector
// clock if the identifiers are large strings.
func NewWithHistory(context context.Context, init Clock, shortenerName string) (*VClock, error) {
	return newClock(context, init, true, shortenerName, true)
}

// Close releases all resources associated with the VClock instance
func (vc *VClock) Close() error {
	vc.cancel()
	return nil
}

// Set assigns the specified value to the given clock identifier.
// The identifier must not be an empty string, nor can an
// identifier be set more than once
func (vc *VClock) Set(id string, v uint64) error {
	return attemptSendChan(vc.req, &SetInfo{Id: id, Value: v}, vc.resp, errClosedVClock)
}

// Tick increments the clock with the specified identifier.
// An error is raised if the identifier is not found in the vector clock
func (vc *VClock) Tick(id string) error {
	return attemptSendChan(vc.req, &reqTick{id: id}, vc.resp, errClosedVClock)
}

// Get returns the latest clock value for the specified identifier,
// returning true if the identifier is found, otherwise false
func (vc *VClock) Get(id string) (uint64, bool) {
	resp, err := attemptSendChanWithResp[*reqGet, *respGetterWithStatus](vc.req, &reqGet{id: id}, vc.resp, errClosedVClock)
	if err != nil {
		return 0, false
	}
	return resp.v, resp.b
}

// GetClock returns a copy of the complete vector clock map
func (vc *VClock) GetClock() (Clock, error) {
	resp, err := attemptSendChanWithResp[*reqSnap, *respClock](vc.req, &reqSnap{}, vc.resp, errClosedVClock)
	if err != nil {
		return nil, err
	}
	if resp.e != nil {
		return nil, resp.e
	}
	return resp.c, nil
}

// GetFullHistory returns a copy of each state change of the vectory clock map,
// including the Event detail of the change as well as new state of the clock
func (vc *VClock) GetFullHistory() ([]*HistoryItem, error) {
	resp, err := attemptSendChanWithResp[*reqFullHistory, *respHistoryAll](vc.req, &reqFullHistory{}, vc.resp, errClosedVClock)
	if err != nil {
		return nil, err
	}
	if resp.e != nil {
		return nil, resp.e
	}
	return resp.h, nil
}

// GetHistory returns a copy of each state change of the vector clock map
func (vc *VClock) GetHistory() ([]Clock, error) {
	resp, err := attemptSendChanWithResp[*reqHistory, *respHistory](vc.req, &reqHistory{}, vc.resp, errClosedVClock)
	if err != nil {
		return nil, err
	}
	if resp.e != nil {
		return nil, resp.e
	}
	return resp.h, nil
}

// Copy creates a new VClock instance, initialised to the
// values of this instance
func (vc *VClock) Copy() (*VClock, error) {
	m, err := vc.GetClock()
	if err != nil {
		return nil, err
	}
	return New(vc.ctx, m, vc.shortener)
}

// LastUpdate returns the latest clock time and its associated identifier
func (vc *VClock) LastUpdate() (string, uint64, error) {
	g, err := attemptSendChanWithResp[*reqLastUpdate, *respGetter](vc.req, &reqLastUpdate{}, vc.resp, errClosedVClock)
	if err != nil {
		return "", 0, err
	}
	if g.e != nil {
		return "", 0, g.e
	}
	return g.id, g.v, nil
}

// Merge combines this clock with the other clock.  The other clock
// must not be nil, and neither must be closed
func (vc *VClock) Merge(other *VClock) error {
	if other == nil {
		return errClockMustNotBeNil
	}

	m, err := other.GetClock()
	if err != nil {
		return err
	}

	return attemptSendChan(vc.req, m, vc.resp, errClosedVClock)
}

// Prune resets the clock history, so that only the latest is available
func (vc *VClock) Prune() error {
	return attemptSendChan(vc.req, &reqPrune{}, vc.resp, errClosedVClock)
}

type clockSerialisation struct {
	B []byte
	C Clock
	S string
}

// Bytes returns an encoded vector clock
func (vc *VClock) Bytes() ([]byte, error) {

	resp, err := attemptSendChanWithResp[*reqSnapShortenedIdentifiers, *respClock](vc.req, &reqSnapShortenedIdentifiers{}, vc.resp, errClosedVClock)
	if err != nil {
		return nil, err
	}
	if resp.e != nil {
		return nil, resp.e
	}

	shortener, err := GetShortenerFactory().Get(vc.shortener)
	if err != nil {
		return nil, err
	}
	b, err := shortener.Bytes()
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)
	if err := enc.Encode(
		&clockSerialisation{
			B: b,
			C: resp.c,
			S: vc.shortener,
		}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// FromBytesWithHistory decodes a vector clock and preserves history from this point forwards.  This requires both
// the serialised clock and also the name of the IdentifierShortener to be used (which may be empty string)
func FromBytesWithHistory(context context.Context, data []byte, shortenerName string) (vc *VClock, err error) {
	return fromBytes(context, data, true, shortenerName)
}

// FromBytes decodes a vector clock.  This requires both
// the serialised clock and also the name of the IdentifierShortener to be used (which may be empty string)
func FromBytes(context context.Context, data []byte, shortenerName string) (vc *VClock, err error) {
	return fromBytes(context, data, false, shortenerName)
}

// fromBytes deseralises and initialises a VClock
func fromBytes(context context.Context, data []byte, maintainHistory bool, shortenerName string) (vc *VClock, err error) {
	b := new(bytes.Buffer)
	b.Write(data)
	dec := gob.NewDecoder(b)

	cs := clockSerialisation{
		C: Clock{},
	}

	if err := dec.Decode(&cs); err != nil {
		return nil, err
	}

	// Retriever the desired shortener
	if shortenerName == "" {
		shortenerName = getDefaultShortenerName()
	}
	sourceShortener, err := GetShortenerFactory().Get(cs.S)
	if err != nil {
		return nil, err
	}

	// As the Clock was serialised using shortened identifiers,
	// if the preferred shortener name differs from that used by the serialising
	// clock, then need to recover to the unshortened identifiers
	if cs.S != shortenerName {
		newC := Clock{}
		for k, v := range cs.C {
			kk, err := sourceShortener.Recover(k)
			if err != nil {
				return nil, err
			}
			newC[kk] = v
		}

		return newClock(context, newC, maintainHistory, shortenerName, true)
	}

	// The two clocks are using the same shortener, we now need to ensure the shortener
	// instance in this process has the same set of mappings as the source of the
	// process whose clock is being merged.
	err = sourceShortener.Merge(cs.B)
	if err != nil {
		return nil, err
	}

	// The new clock can be created successfully, since the shortener now
	// has all necessary mappings to be able to fully recover the original identifiers
	// for all entries in the clock, without needing a central service.
	return newClock(context, cs.C, maintainHistory, shortenerName, false)
}

// Compare takes another clock and determines if it is Equal, an
// Ancestor, Descendant, or Concurrent with the callees clock.
func (vc *VClock) compare(other *VClock, cond condition) (bool, error) {
	if other == nil {
		return false, errClockMustNotBeNil
	}

	m, err := other.GetClock()
	if err != nil {
		return false, err
	}

	return attemptSendChanWithResp[*respComp, bool](vc.req, &respComp{other: m, cond: cond}, vc.resp, errClosedVClock)
}

// Equal returns true if the contents of the other clock
// exactly match this instance.
func (vc *VClock) Equal(other *VClock) (bool, error) {
	return vc.compare(other, equal)
}

// Concurrent returns true if the contents of the other clock
// are either completely or partially distinct.  Where partially
// distinct, matching identifiers in the clocks must have the same value.
func (vc *VClock) Concurrent(other *VClock) (bool, error) {
	return vc.compare(other, concurrent)
}

// DescendsFrom returns true if the contents of the other clock shows
// that it can have descended from this clock instance.  This means
// that this clock's identifiers must all be present in the other clock,
// and that this clock's identifier values must all be the same or
// less than their value in the other clock, with at least one
// identifier's value being less.
func (vc *VClock) DescendsFrom(other *VClock) (bool, error) {
	return vc.compare(other, descendant)
}

// AncestorOf returns true if the contents of this clock instance shows
// that it can have descended from the other clock instance.  This means
// that the other clock's identifiers must all be present in the this clock,
// and that the other clock's identifier values must all be the same or
// less than their value in this clock, with at least one of the other clock's
// identifier's value being less.
func (vc *VClock) AncestorOf(other *VClock) (bool, error) {
	return vc.compare(other, ancestor)
}

func getDefaultShortenerName() string {
	return "NoOp"
}

// newClock starts a new clock, with or without history
func newClock(ctx context.Context, init Clock, maintainHistory bool, shortenerName string, applyShortenerToInit bool) (*VClock, error) {

	ctx, cancel := context.WithCancel(ctx)

	v := &VClock{
		req:       chant.New[any](),
		resp:      chant.New[any](),
		shortener: shortenerName,
		ctx:       ctx,
		cancel:    cancel,
	}

	if v.shortener == "" {
		v.shortener = getDefaultShortenerName()
	}
	shortener, _ := GetShortenerFactory().Get(v.shortener)

	waiter := make(chan bool)

	go func() {

		defer func() {
			v.req.Close()
			v.resp.Close()
		}()

		noErr := &respErr{err: nil}

		c := Clock{}
		if init != nil {
			keys := syncmap.SortedKeys(init)
			for _, key := range keys {
				c[key] = init[key]
			}
		}

		history := newHistory(c, shortener, applyShortenerToInit)

		processRequest := func(r any) {

			if !maintainHistory {
				// Prune if history not being maintained
				history = newHistory(history.latest(), shortener, false)
			}

			switch t := r.(type) {
			case *respComp:
				{
					f := func(s string) (string, error) { return shortener.Shorten(s), nil }
					c, _ := copyMapWithKeyModification(t.other, f)
					v.resp.Send(compare(history.latest(), c, t.cond))
				}
			case *reqFullHistory:
				{
					h, err := history.getFullAll()
					v.resp.Send(&respHistoryAll{h: h, e: err})
				}
			case *reqGet:
				{
					vc := history.latest()

					val, ok := vc[shortener.Shorten(t.id)]
					g := &respGetterWithStatus{b: ok}
					g.id = t.id
					g.v = val
					v.resp.Send(g)
				}
			case *reqHistory:
				{
					h, err := history.getAll()
					v.resp.Send(&respHistory{h: h, e: err})
				}
			case *reqLastUpdate:
				{
					vc := history.latest()

					var id string = ""
					var last uint64
					for key := range vc {
						if vc[key] > last {
							id = key
							last = vc[key]
						}
					}
					id, err := shortener.Recover(id)
					v.resp.Send(&respGetter{id: id, v: last, e: err})
				}
			case Clock:
				{
					v.resp.Send(&respErr{err: history.apply(&Event{Type: Merge, Merge: t})})
				}
			case *reqPrune:
				{
					history = newHistory(history.latest(), shortener, false)
					v.resp.Send(noErr)
				}
			case *SetInfo:
				{
					v.resp.Send(&respErr{err: history.apply(&Event{Type: Set, Set: t})})
				}
			case *reqSnap:
				{
					c, err := history.latestWithCopy(false)
					v.resp.Send(&respClock{c: c, e: err})
				}
			case *reqSnapShortenedIdentifiers:
				{
					c, err := history.latestWithCopy(true)
					v.resp.Send(&respClock{c: c, e: err})
				}
			case *reqTick:
				{
					if len(t.id) == 0 {
						v.resp.Send(&respErr{err: errClockIdMustNotBeEmptyString})
					} else {
						v.resp.Send(&respErr{err: history.apply(&Event{Type: Tick, Tick: t.id})})
					}
				}
			default:
				v.resp.Send(&respErr{err: errUnknownReqType})
			}
		}

		// Signal ready
		waiter <- true

		for {
			select {
			case <-ctx.Done():
				return
			case r := <-v.req.RawChan():
				processRequest(r)
			}
		}

	}()

	// Wait until ready
	<-waiter
	close(waiter)

	return v, nil
}
