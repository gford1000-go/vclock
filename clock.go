package vclock

import (
	"bytes"
	"cmp"
	"encoding/gob"
	"errors"
	"sort"
)

// Clock is the underlying type of the vector clock
type Clock map[string]uint64

type AllowedReq interface {
	Clock | *respComp | *reqEnd | *reqFullHistory | *reqGet | *reqHistory | *reqLastUpdate | *reqPrune | *reqSnap | *SetInfo | *reqTick
}

type AllowedResp interface {
	Clock | []Clock | *respErr | bool | *respGetter | *respGetterWithStatus | []*HistoryItem
}

// attemptSendChanWithResp will stop the panic and return recoverErr, should the chan be closed
func attemptSendChanWithResp[T AllowedReq, U AllowedResp](c chan any, t T, r chan any, recoverErr error) (u U, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverErr
		}
	}()
	c <- t
	a := <-r
	return a.(U), nil
}

// attemptSendChan is syntax sugar to simply the call when only an error would be returned
func attemptSendChan[T AllowedReq](c chan any, t T, r chan any, recoverErr error) error {
	resp, err := attemptSendChanWithResp[T, *respErr](c, t, r, recoverErr)
	if err != nil {
		return err
	}
	return resp.err
}

// sortedKeys returns a sorted slice of the map's keys
func sortedKeys[K cmp.Ordered, V any](m map[K]V) []K {
	keys := make([]K, len(m))
	i := 0
	for k := range m {
		keys[i] = k
		i++
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}

type reqEnd struct {
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

type reqTick struct {
	id string
}

type respErr struct {
	err error
}

type respGetter struct {
	id string
	v  uint64
}

type respGetterWithStatus struct {
	respGetter
	b bool
}

var errClockIdMustNotBeEmptyString = errors.New("clock identifier must not be empty string")
var errAttemptToSetExistingId = errors.New("clock identifier cannot be reset once initialised")
var errAttemptToTickUnknownId = errors.New("attempted to tick unknown clock identifier")
var errClosedVClock = errors.New("attempt to interact with closed clock")
var errClockMustNotBeNil = errors.New("attempt to merge a nil clock")

// VClock is an instance of a vector clock that can suppport
// concurrent use across multiple goroutines
type VClock struct {
	req  chan any
	resp chan any
}

// New returns a VClock initialised with the specified pairs,
// which does not maintain any history.
func New(init map[string]uint64) (*VClock, error) {
	return newClock(init, false)
}

// NewWithHistory returns a VClock initialised with the specified
// pairs, which maintains a full history.
func NewWithHistory(init map[string]uint64) (*VClock, error) {
	return newClock(init, true)
}

// Close releases all resources associated with the VClock instance
func (vc *VClock) Close() error {
	return attemptSendChan(vc.req, &reqEnd{}, vc.resp, errClosedVClock)
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

// GetMap returns a copy of the complete vector clock map
func (vc *VClock) GetMap() (Clock, error) {
	return attemptSendChanWithResp[*reqSnap, Clock](vc.req, &reqSnap{}, vc.resp, errClosedVClock)
}

// GetFullHistory returns a copy of each state change of the vectory clock map,
// including the Event detail of the change as well as new state of the clock
func (vc *VClock) GetFullHistory() ([]*HistoryItem, error) {
	return attemptSendChanWithResp[*reqFullHistory, []*HistoryItem](vc.req, &reqFullHistory{}, vc.resp, errClosedVClock)
}

// GetHistory returns a copy of each state change of the vector clock map
func (vc *VClock) GetHistory() ([]Clock, error) {
	return attemptSendChanWithResp[*reqHistory, []Clock](vc.req, &reqHistory{}, vc.resp, errClosedVClock)
}

// Copy creates a new VClock instance, initialised to the
// values of this instance
func (vc *VClock) Copy() (*VClock, error) {
	m, err := vc.GetMap()
	if err != nil {
		return nil, err
	}
	return New(m)
}

// LastUpdate returns the latest clock time and its associated identifier
func (vc *VClock) LastUpdate() (id string, last uint64) {
	g, err := attemptSendChanWithResp[*reqLastUpdate, *respGetter](vc.req, &reqLastUpdate{}, vc.resp, errClosedVClock)
	if err != nil {
		return "", 0
	}
	return g.id, g.v
}

// Merge combines this clock with the other clock.  The other clock
// must not be nil, and neither must be closed
func (vc *VClock) Merge(other *VClock) error {
	if other == nil {
		return errClockMustNotBeNil
	}

	m, err := other.GetMap()
	if err != nil {
		return err
	}

	return attemptSendChan(vc.req, m, vc.resp, errClosedVClock)
}

// Prune resets the clock history, so that only the latest is available
func (vc *VClock) Prune() error {
	return attemptSendChan(vc.req, &reqPrune{}, vc.resp, errClosedVClock)
}

// Bytes returns an encoded vector clock
func (vc *VClock) Bytes() ([]byte, error) {
	m, err := vc.GetMap()
	if err != nil {
		return nil, err
	}

	b := new(bytes.Buffer)
	enc := gob.NewEncoder(b)
	if err := enc.Encode(m); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// FromBytes decodes a vector clock
func FromBytes(data []byte) (vc *VClock, err error) {
	b := new(bytes.Buffer)
	b.Write(data)
	dec := gob.NewDecoder(b)

	m := map[string]uint64{}
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}
	return New(m)
}

// Compare takes another clock and determines if it is Equal, an
// Ancestor, Descendant, or Concurrent with the callees clock.
func (vc *VClock) compare(other *VClock, cond condition) (bool, error) {
	if other == nil {
		return false, errClockMustNotBeNil
	}

	m, err := other.GetMap()
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

// newClock starts a new clock, with or without history
func newClock(init Clock, maintainHistory bool) (*VClock, error) {
	v := &VClock{
		req:  make(chan any),
		resp: make(chan any),
	}

	go clockLoop(v, init, maintainHistory)

	return v, nil
}

// clockLoop is the goroutine started within calls to New...
func clockLoop(v *VClock, init Clock, maintainHistory bool) {
	defer func() {
		close(v.resp)
	}()

	noErr := &respErr{err: nil}

	c := Clock{}
	if init != nil {
		keys := sortedKeys(init)
		for _, key := range keys {
			c[key] = init[key]
		}
	}

	history := newHistory(c)

	for r := range v.req {

		if !maintainHistory {
			// Prune if history not being maintained
			history = newHistory(history.latest())
		}

		switch t := r.(type) {
		case *reqEnd:
			{
				close(v.req)
				v.resp <- noErr
			}
		case *respComp:
			{
				v.resp <- compare(history.latest(), t.other, t.cond)
			}
		case *reqFullHistory:
			{
				v.resp <- history.getFullAll()
			}
		case *reqGet:
			{
				vc := history.latest()

				val, ok := vc[t.id]
				g := &respGetterWithStatus{b: ok}
				g.id = t.id
				g.v = val
				v.resp <- g
			}
		case *reqHistory:
			{
				v.resp <- history.getAll()
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
				v.resp <- &respGetter{id: id, v: last}
			}
		case Clock:
			{
				v.resp <- &respErr{err: history.apply(&Event{Type: Merge, Merge: t})}
			}
		case *reqPrune:
			{
				history = newHistory(history.latest())
				v.resp <- noErr
			}
		case *SetInfo:
			{
				v.resp <- &respErr{err: history.apply(&Event{Type: Set, Set: t})}
			}
		case *reqSnap:
			{
				v.resp <- history.latestWithCopy()
			}
		case *reqTick:
			{
				if len(t.id) == 0 {
					v.resp <- &respErr{err: errClockIdMustNotBeEmptyString}
				} else {
					v.resp <- &respErr{err: history.apply(&Event{Type: Tick, Tick: t.id})}
				}
			}
		}
	}
}
