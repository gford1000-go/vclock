package vclock

// copyMap returns a copy of the supplied instance (non-deep)
func copyMap[T comparable, U any](m map[T]U) map[T]U {
	newm := map[T]U{}
	for k, v := range m {
		newm[k] = v
	}
	return newm
}

// history records all historic events (subject to pruning)
type history struct {
	lastId uint64
	items  []*HistoryItem
}

// apply attempts to extend the history by applying the event
func (h *history) apply(event *Event) error {
	vc := h.latestWithCopy()
	if err := event.apply(vc); err != nil {
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

// latest returns the current clock value
func (h *history) latest() Clock {
	return h.items[h.getLastId()].Clock
}

// latestWithCopy returns a copy of the current clock value
func (h *history) latestWithCopy() Clock {
	return copyMap(h.latest())
}

// getLastId returns the id of the latest clock
func (h *history) getLastId() uint64 {
	return h.lastId
}

// getRange returns the specified range of history
func (h *history) getRange(from, to uint64) []Clock {
	if from > to {
		return h.getRange(to, from)
	}
	ret := []Clock{}
	for i := from; i <= to; i++ {
		if i <= h.getLastId() {
			ret = append(ret, copyMap(h.items[i].Clock))
		}
	}
	return ret
}

// getAll returns all of the history
func (h *history) getAll() []Clock {
	return h.getRange(0, h.getLastId())
}

// getFullRange returns the specified range of history
func (h *history) getFullRange(from, to uint64) []*HistoryItem {
	if from > to {
		return h.getFullRange(to, from)
	}
	ret := []*HistoryItem{}
	for i := from; i <= to; i++ {
		if i <= h.getLastId() {
			ret = append(ret, h.items[i].copy())
		}
	}
	return ret
}

// getFullAll returns all of the history
func (h *history) getFullAll() []*HistoryItem {
	return h.getFullRange(0, h.getLastId())
}

// newHistory initialises an instance of history
func newHistory(m Clock) *history {
	return &history{
		lastId: 0,
		items: []*HistoryItem{
			{
				HistoryId: 0,
				Change:    nil,
				Clock:     copyMap(m),
			},
		},
	}
}
