package vclock

import (
	"bytes"
	"encoding/gob"
	"errors"

	"github.com/gford1000-go/syncmap"
)

// ShortenedMap is the underlying type expected to be serialised by IdentifierShortener implementations
type ShortenedMap map[string]string

// IdentifierShortener provides functions to shorten vector clock
// identifiers to minimise the overall memory footprint of the clock.
type IdentifierShortener interface {
	Name() string            // Name of the shortener - msut be unique
	Shorten(s string) string // Returns the shortened version of the supplied string
	Recover(s string) string // Recovers the original string from the shortened version
	Bytes() ([]byte, error)  // The full map of shortened strings to original strings as a serialised ShortenedMap
	Merge(b []byte) error    // Merge the contents of the ShortenedMap into the instance
}

// Shortener is the function that applies the transformation
type Shortener func(string) string

var errShortenerIsNil = errors.New("shortener must not be nil")
var errShortenerNameIsNil = errors.New("shortener name must be non-empty string")
var errSerialiseNameMismatch = errors.New("shortener name mismatch - deserialisation not possible")

// NewInMemoryShortener creates an instance of InMemoryShortener that will
// use the specified Shortener
func NewInMemoryShortener(name string, shortener Shortener) (*InMemoryShortener, error) {
	if len(name) == 0 {
		return nil, errShortenerNameIsNil
	}
	if shortener == nil {
		return nil, errShortenerIsNil
	}

	return &InMemoryShortener{
		sm: syncmap.New[string, string](nil),
		f:  shortener,
		n:  name,
	}, nil
}

// InMemoryShortener uses a map to store the results
// of Shorten for a given string, so that it can be
// easily recovered.
type InMemoryShortener struct {
	sm *syncmap.SynchronisedMap[string, string]
	f  Shortener
	n  string
}

func (h *InMemoryShortener) Name() string {
	return h.n
}

func (h *InMemoryShortener) Shorten(s string) string {
	k := h.f(s)
	h.sm.Insert(k, s, false)
	return k
}

func (h *InMemoryShortener) Recover(s string) string {
	if ss, err := h.sm.Get(s); err != nil {
		return ""
	} else {
		return ss
	}
}

type serial struct {
	N string
	B []byte
}

func (h *InMemoryShortener) Bytes() ([]byte, error) {
	b, err := h.sm.Bytes()
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	enc := gob.NewEncoder(buf)

	s := serial{
		N: h.Name(),
		B: b,
	}

	if err := enc.Encode(&s); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (h *InMemoryShortener) Merge(b []byte) error {

	buf := new(bytes.Buffer)
	buf.Write(b)
	dec := gob.NewDecoder(buf)

	s := serial{
		B: []byte{},
	}

	if err := dec.Decode(&s); err != nil {
		return err
	}

	if s.N != h.Name() {
		return errSerialiseNameMismatch
	}

	return h.sm.Merge(s.B)
}
