package vclock

import (
	"errors"
)

// IdentifierShortener provides functions to shorten vector clock
// identifiers to minimise the overall memory footprint of the clock.
type IdentifierShortener interface {
	Name() string
	Shorten(s string) string
	Recover(s string) string
}

// Shortener is the function that applies the transformation
type Shortener func(string) string

var errShortenerIsNil = errors.New("shortener must not be nil")
var errShortenerNameIsNil = errors.New("shortener name must be non-empty string")

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
		sm: NewSynchronisedMap[string, string](nil),
		f:  shortener,
		n:  name,
	}, nil
}

// InMemoryShortener uses a map to store the results
// of Shorten for a given string, so that it can be
// easily recovered.
type InMemoryShortener struct {
	sm *SynchronisedMap[string, string]
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
