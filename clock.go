package vclock

import (
	"bytes"
	"encoding/gob"
	"errors"
)

// attemptSendChan will stop the panic and return recoverErr, should the chan be closed
func attemptSendChan[T interface{}](c chan T, i T, e chan error, recoverErr error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverErr
		}
	}()
	c <- i
	return <-e
}

// attemptSendChanWithResp will stop the panic and return recoverErr, should the chan be closed
func attemptSendChanWithResp[T interface{}, U interface{}](c chan T, i T, r chan U, recoverErr error) (u U, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = recoverErr
		}
	}()
	c <- i
	return <-r, nil
}

// Condition constants define how to compare a vector clock against another,
// and may be ORed together when being provided to the Compare method.
type Condition int

// Constants define compairison conditions between pairs of vector clocks
const (
	Equal Condition = 1 << iota
	Ancestor
	Descendant
	Concurrent
)

type comp struct {
	other map[string]uint64
	cond  Condition
}

type getter struct {
	id string
	v  uint64
}

type getterWithStatus struct {
	getter
	b bool
}

type setter struct {
	id string
	v  uint64
}

var errClockIdMustNotBeEmptyString = errors.New("clock identifier must not be empty string")
var errAttemptToSetExistingId = errors.New("clock identifier cannot be reset once initialised")
var errAttemptToTickUnknownId = errors.New("attempted to tick unknown clock identifier")
var errClosedVClock = errors.New("attempt to interact with closed clock")
var errClockMustNotBeNil = errors.New("attempt to merge a nil clock")
var errEnded = errors.New("ended clock")

// VClock is an instance of a vector clock that can suppport
// concurrent use across multiple goroutines
type VClock struct {
	end            chan bool
	err            chan error
	reqComp        chan *comp
	respComp       chan bool
	reqGet         chan string
	respGet        chan *getterWithStatus
	reqLastUpdate  chan bool
	respLastUpdate chan *getter
	reqMerge       chan map[string]uint64
	reqSnap        chan bool
	respSnap       chan map[string]uint64
	reqTick        chan string
	setter         chan *setter
}

// New returns a VClock initialised with the specified pairs.
func New(init map[string]uint64) (*VClock, error) {
	v := &VClock{
		end:            make(chan bool),
		err:            make(chan error),
		reqComp:        make(chan *comp),
		respComp:       make(chan bool),
		reqGet:         make(chan string),
		respGet:        make(chan *getterWithStatus),
		reqLastUpdate:  make(chan bool),
		respLastUpdate: make(chan *getter),
		reqMerge:       make(chan map[string]uint64),
		reqSnap:        make(chan bool),
		respSnap:       make(chan map[string]uint64),
		reqTick:        make(chan string),
		setter:         make(chan *setter),
	}

	go func() {
		defer func() {
			close(v.end)
			close(v.err)
		}()

		closer := func() {
			close(v.reqComp)
			close(v.respComp)
			close(v.reqGet)
			close(v.respGet)
			close(v.reqLastUpdate)
			close(v.respLastUpdate)
			close(v.reqMerge)
			close(v.reqSnap)
			close(v.respSnap)
			close(v.reqTick)
			close(v.setter)
		}

		var vc = make(map[string]uint64)

		for {
			select {
			case <-v.end:
				closer() // Prevent any further attempts to make requests
				v.err <- errEnded
				return
			case c := <-v.reqComp:
				{
					// Compare takes another clock and determines if it is Equal, an
					// Ancestor, Descendant, or Concurrent with the callees clock.

					var otherIs Condition
					// Preliminary qualification based on length
					if len(vc) > len(c.other) {
						if c.cond&(Ancestor|Concurrent) == 0 {
							v.respComp <- false
							continue
						}
						otherIs = Ancestor
					} else if len(vc) < len(c.other) {
						if c.cond&(Descendant|Concurrent) == 0 {
							v.respComp <- false
							continue
						}
						otherIs = Descendant
					} else {
						otherIs = Equal
					}

					// Compare matching items
					applyTest := true
					for id := range c.other {
						if _, found := vc[id]; found {
							if c.other[id] > vc[id] {
								switch otherIs {
								case Equal:
									if c.cond&Descendant == 0 {
										v.respComp <- false
										applyTest = false
										goto checkContinue
									}
									otherIs = Descendant
								case Ancestor:
									v.respComp <- c.cond&Concurrent != 0
									applyTest = false
								}
							} else if c.other[id] < vc[id] {
								switch otherIs {
								case Equal:
									if c.cond&Ancestor == 0 {
										v.respComp <- false
										applyTest = false
										goto checkContinue
									}
									otherIs = Ancestor
								case Descendant:
									v.respComp <- c.cond&Concurrent != 0
									applyTest = false
								}
							}
						} else {
							if otherIs == Equal {
								v.respComp <- c.cond&Concurrent != 0
								applyTest = false
							} else if (len(c.other) - len(vc) - 1) < 0 {
								v.respComp <- c.cond&Concurrent != 0
								applyTest = false
							}
						}
					}
				checkContinue:
					if !applyTest {
						break
					}
					v.respComp <- c.cond&otherIs != 0
				}
			case id := <-v.reqGet:
				{
					val, ok := vc[id]
					g := &getterWithStatus{b: ok}
					g.id = id
					g.v = val
					v.respGet <- g
				}
			case <-v.reqLastUpdate:
				{
					var id string = ""
					var last uint64
					for key := range vc {
						if vc[key] > last {
							id = key
							last = vc[key]
						}
					}
					v.respLastUpdate <- &getter{id: id, v: last}
				}
			case other := <-v.reqMerge:
				{
					for id := range other {
						if _, ok := vc[id]; ok {
							if vc[id] < other[id] {
								vc[id] = other[id]
							}
						} else {
							vc[id] = other[id]
						}
					}
					v.err <- nil
				}
			case <-v.reqSnap:
				{
					m := map[string]uint64{}
					for key, val := range vc {
						m[key] = val
					}
					v.respSnap <- m
				}
			case p := <-v.setter:
				{
					if len(p.id) == 0 {
						v.err <- errClockIdMustNotBeEmptyString
					} else {
						if _, ok := vc[p.id]; !ok {
							vc[p.id] = p.v
							v.err <- nil
						} else {
							v.err <- errAttemptToSetExistingId
						}
					}
				}
			case id := <-v.reqTick:
				{
					if len(id) == 0 {
						v.err <- errClockIdMustNotBeEmptyString
					} else {
						if _, ok := vc[id]; ok {
							vc[id] += 1
							v.err <- nil
						} else {
							v.err <- errAttemptToTickUnknownId
						}
					}
				}
			}
		}

	}()

	for key, val := range init {
		if err := v.Set(key, val); err != nil {
			return nil, v.Close()
		}
	}

	return v, nil
}

// Close releases all resources associated with the VClock instance
func (vc *VClock) Close() error {
	return attemptSendChan(vc.end, true, vc.err, errClosedVClock)
}

// Set assigns the specified value to the given clock identifier.
// The identifier must not be an empty string, nor can an
// identifier be set more than once
func (vc *VClock) Set(id string, v uint64) error {
	return attemptSendChan(vc.setter, &setter{id: id, v: v}, vc.err, errClosedVClock)
}

// Tick increments the clock with the specified identifier.
// An error is raised if the identifier is not found in the vector clock
func (vc *VClock) Tick(id string) error {
	return attemptSendChan(vc.reqTick, id, vc.err, errClosedVClock)
}

// Get returns the latest clock value for the specified identifier,
// returning true if the identifier is found, otherwise false
func (vc *VClock) Get(id string) (uint64, bool) {
	resp, err := attemptSendChanWithResp(vc.reqGet, id, vc.respGet, errClosedVClock)
	if err != nil {
		return 0, false
	}
	return resp.v, resp.b
}

// GetMap returns a copy of the complete vector clock map
func (vc *VClock) GetMap() (map[string]uint64, error) {
	return attemptSendChanWithResp(vc.reqSnap, true, vc.respSnap, errClosedVClock)
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
	g, err := attemptSendChanWithResp(vc.reqLastUpdate, true, vc.respLastUpdate, errClosedVClock)
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

	return attemptSendChan(vc.reqMerge, m, vc.err, errClosedVClock)
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
func (vc *VClock) Compare(other *VClock, cond Condition) (bool, error) {
	if other == nil {
		return false, errClockMustNotBeNil
	}

	m, err := other.GetMap()
	if err != nil {
		return false, err
	}

	return attemptSendChanWithResp(vc.reqComp, &comp{other: m, cond: cond}, vc.respComp, errClosedVClock)
}
