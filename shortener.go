package vclock

// IdentifierShortener provides functions to shorten vector clock
// identifiers to minimise the overall memory footprint of the clock.
type IdentifierShortener interface {
	Shorten(s string) string
	Recover(s string) string
}

// noopShortener does nothing, and is used when no IdentifierShortener
// is provided when calling New
type noopShortener struct {
}

func (n *noopShortener) Shorten(s string) string {
	return s
}

func (n *noopShortener) Recover(s string) string {
	return s
}
