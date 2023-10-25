package vclock

// copyMap returns a copy of the supplied instance (non-deep)
func copyMap[T comparable, U any](m map[T]U) map[T]U {
	newm := map[T]U{}
	for k, v := range m {
		newm[k] = v
	}
	return newm
}

// event updates the vector clock
type event struct {
	setter *setter
	tick   string
	merge  map[string]uint64
}

// apply attempts to assign the change to the supplied map
func (e *event) apply(m map[string]uint64) error {
	if e.setter != nil {
		if len(e.setter.id) == 0 {
			return errClockIdMustNotBeEmptyString
		} else {
			if _, ok := m[e.setter.id]; !ok {
				m[e.setter.id] = e.setter.v
			} else {
				return errAttemptToSetExistingId
			}
		}
	} else if len(e.tick) > 0 {
		if _, ok := m[e.tick]; ok {
			m[e.tick] += 1
		} else {
			return errAttemptToTickUnknownId
		}
	} else if e.merge != nil {
		for id := range e.merge {
			if _, ok := m[id]; ok {
				if m[id] < e.merge[id] {
					m[id] = e.merge[id]
				}
			} else {
				m[id] = e.merge[id]
			}
		}
	}
	return nil
}

// item records the new state after an event is applied
type item struct {
	id    uint64
	event *event
	vc    map[string]uint64
}

// history records all historic events (subject to pruning)
type history struct {
	lastId uint64
	items  []*item
}

// apply attempts to extend the history by applying the event
func (h *history) apply(event *event) error {
	vc := h.latestWithCopy()
	if err := event.apply(vc); err != nil {
		return err
	}

	nextId := h.getLastId() + 1

	item := &item{
		id:    nextId,
		event: event,
		vc:    vc,
	}

	h.items = append(h.items, item)
	h.lastId = nextId
	return nil
}

// latest returns the current clock value
func (h *history) latest() map[string]uint64 {
	return h.items[h.getLastId()].vc
}

// latestWithCopy returns a copy of the current clock value
func (h *history) latestWithCopy() map[string]uint64 {
	return copyMap(h.latest())
}

// getLastId returns the id of the latest clock
func (h *history) getLastId() uint64 {
	return h.lastId
}

// getRange returns the specified range of history
func (h *history) getRange(from, to uint64) []map[string]uint64 {
	if from > to {
		return h.getRange(to, from)
	}
	ret := []map[string]uint64{}
	for i := from; i <= to; i++ {
		if i <= h.getLastId() {
			ret = append(ret, copyMap(h.items[i].vc))
		}
	}
	return ret
}

// getAll returns all of the history
func (h *history) getAll() []map[string]uint64 {
	return h.getRange(0, h.getLastId())
}

// newHistory initialises an instance of history
func newHistory(m map[string]uint64) *history {
	return &history{
		lastId: 0,
		items: []*item{
			{
				id:    0,
				event: nil,
				vc:    copyMap(m),
			},
		},
	}
}
