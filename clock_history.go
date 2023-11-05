package vclock

// copyMap returns a copy of the supplied instance (non-deep)
func copyMap[T comparable, U any](m map[T]U) map[T]U {
	newm := map[T]U{}
	for k, v := range m {
		newm[k] = v
	}
	return newm
}

// copyMapWithKeyModification applies the function to the key as it is copied
func copyMapWithKeyModification[T comparable, U any](m map[T]U, f func(T) (T, error)) (map[T]U, error) {
	newm := map[T]U{}
	for k, v := range m {
		kk, err := f(k)
		if err != nil {
			return nil, err
		}
		newm[kk] = v
	}
	return newm, nil
}

// history records all historic events (subject to pruning)
// all HistoryItems contain Clocks with shortened identifiers,
// which are created during the apply().
type history struct {
	lastId    uint64
	items     []*HistoryItem
	shortener IdentifierShortener
}

// apply attempts to extend the history by applying the event
func (h *history) apply(event *Event) error {
	vc, err := h.latestWithCopy(true)
	if err != nil {
		return err
	}

	if err := event.apply(vc, h.shortener.Shorten); err != nil {
		return err
	}

	nextId := h.getLastId() + 1

	item := &HistoryItem{
		HistoryId: nextId,
		Change:    event,
		Clock:     vc,
	}

	h.items = append(h.items, item)
	h.lastId = nextId
	return nil
}

// latest returns the current clock value unaltered
// i.e. always with the shortened identifiers
func (h *history) latest() Clock {
	return h.items[h.getLastId()].Clock
}

// latestWithCopy returns a copy of the current clock value,
// with either shortened or full identifiers
func (h *history) latestWithCopy(useExistingIdentifiers bool) (Clock, error) {
	if useExistingIdentifiers {
		return copyMap(h.latest()), nil
	}
	return copyMapWithKeyModification[string, uint64](h.latest(), h.shortener.Recover)
}

// getLastId returns the id of the latest clock
func (h *history) getLastId() uint64 {
	return h.lastId
}

// getRange returns the specified range of history
func (h *history) getRange(from, to uint64, useShortened bool) ([]Clock, error) {
	if from > to {
		return h.getRange(to, from, useShortened)
	}
	ret := []Clock{}
	for i := from; i <= to; i++ {
		if i <= h.getLastId() {
			if useShortened {
				ret = append(ret, copyMap(h.items[i].Clock))
			} else {
				m, err := copyMapWithKeyModification(h.items[i].Clock, h.shortener.Recover)
				if err != nil {
					return nil, err
				}
				ret = append(ret, m)
			}
		}
	}
	return ret, nil
}

// getAll returns all of the history using the
// fully expanded identifiers
func (h *history) getAll() ([]Clock, error) {
	return h.getRange(0, h.getLastId(), false)
}

// getFullRange returns the specified range of history
func (h *history) getFullRange(from, to uint64, useShortened bool) ([]*HistoryItem, error) {
	if from > to {
		return h.getFullRange(to, from, useShortened)
	}
	ret := []*HistoryItem{}
	for i := from; i <= to; i++ {
		if i <= h.getLastId() {
			if useShortened {
				ret = append(ret, h.items[i].copy())
			} else {
				item, err := h.items[i].copyWithKeyModification(h.shortener.Recover)
				if err != nil {
					return nil, err
				}
				ret = append(ret, item)
			}
		}
	}
	return ret, nil
}

// getFullAll returns all of the history using the
// fully expanded identifiers
func (h *history) getFullAll() ([]*HistoryItem, error) {
	return h.getFullRange(0, h.getLastId(), false)
}

// newHistory initialises an instance of history
func newHistory(m Clock, shortener IdentifierShortener, applyShortener bool) *history {
	h := &history{
		lastId:    0,
		items:     []*HistoryItem{},
		shortener: shortener,
	}

	var c Clock
	if applyShortener {
		f := func(s string) (string, error) { return shortener.Shorten(s), nil }
		c, _ = copyMapWithKeyModification(m, f)
	} else {
		c = copyMap(m)
	}

	h.items = append(h.items, &HistoryItem{
		HistoryId: 0,
		Change:    nil,
		Clock:     c,
	})

	return h
}
