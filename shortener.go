package vclock

import (
	"errors"
	"sync"
)

// IdentifierShortener provides functions to shorten vector clock
// identifiers to minimise the overall memory footprint of the clock.
type IdentifierShortener interface {
	Shorten(s string) string
	Recover(s string) string
}

// Shortener is the function that applies the transformation
type Shortener func(string) string

var errShortenerIsNil = errors.New("shortener must not be nil")

// NewShortener creates an instance of InMemoryShortener that will
// use the specified Shortener
func NewShortener(shortener Shortener) (*InMemoryShortener, error) {
	if shortener == nil {
		return nil, errShortenerIsNil
	}

	return &InMemoryShortener{
		m: map[string]string{},
		f: shortener,
	}, nil
}

// InMemoryShortener uses a map to store the results
// of Shorten for a given string, so that it can be
// easily recovered.
type InMemoryShortener struct {
	lck sync.Mutex
	m   map[string]string
	f   Shortener
}

func (h *InMemoryShortener) Shorten(s string) string {
	h.lck.Lock()
	defer h.lck.Unlock()

	k := h.f(s)
	h.m[k] = s
	return k
}

func (h *InMemoryShortener) Recover(s string) string {
	h.lck.Lock()
	defer h.lck.Unlock()
	return h.m[s]
}
