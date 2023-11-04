package vclock

import "errors"

// factory is a singleton instance of a factory
var factory *ShortenerFactory

func init() {
	factory = &ShortenerFactory{
		m: NewSynchronisedMap[string, IdentifierShortener](nil),
	}

	noop, _ := NewInMemoryShortener("NoOp", func(s string) string { return s })
	factory.Register(noop)
}

// GetShortenerFactory returns the ShortenerFactory
func GetShortenerFactory() *ShortenerFactory {
	return factory
}

var ErrShortenerMustNotBeNil = errors.New("shortener cannot be nil")

// ShortenerFactory manages IdentifierShortener instances
type ShortenerFactory struct {
	m *SynchronisedMap[string, IdentifierShortener]
}

// Register adds the specified shortener, returns error if the shortener
// is already registered (i.e. key with the shortener name already exists)
func (f *ShortenerFactory) Register(shortener IdentifierShortener) error {
	if shortener == nil {
		return ErrShortenerMustNotBeNil
	}
	_, err := f.m.Insert(shortener.Name(), shortener, true)
	return err
}

// Names returns the list of shorteners in the factory
func (f *ShortenerFactory) Names() []string {
	return f.m.GetKeys()
}

// Get returns the IdentifierShortener with the specified name, or
// an error if not found.
func (f *ShortenerFactory) Get(name string) (IdentifierShortener, error) {
	return f.m.Get(name)
}
