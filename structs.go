package vclock

import "fmt"

// SetInfo stores the value to be applied to the vector clock
// for the specified identifier.
type SetInfo struct {
	Id    string
	Value uint64
}

func (s *SetInfo) String() string {
	return fmt.Sprint(*s)
}

// copy returns a copy of the instance
func (s *SetInfo) copy() *SetInfo {
	return &SetInfo{Id: s.Id, Value: s.Value}
}

// EventType describes the type of update within an Event
type EventType uint

func (e EventType) String() string {
	switch e {
	case Set:
		return "Set"
	case Tick:
		return "Tick"
	case Merge:
		return "Merge"
	}
	return "Unknown"
}

const (
	Set EventType = 1 << iota
	Tick
	Merge
)

// Event captures the details of a specific update to the vector clock.
// Only one of the attributes will contain information.
type Event struct {
	Type  EventType
	Set   *SetInfo
	Tick  string
	Merge Clock
}

func (e *Event) String() string {
	return fmt.Sprint(*e)
}

// copy returns a deep copy of the instance
func (e *Event) copy() *Event {
	ret := &Event{Type: e.Type}
	switch e.Type {
	case Set:
		ret.Set = e.Set.copy()
	case Tick:
		ret.Tick = e.Tick
	case Merge:
		ret.Merge = copyMap(e.Merge)
	}
	return ret
}

// apply attempts to assign the change to the supplied map,
// transforming the identifiers using the supplied function
func (e *Event) apply(m Clock, f func(string) string) error {
	switch e.Type {
	case Set:
		if len(e.Set.Id) == 0 {
			return errClockIdMustNotBeEmptyString
		}

		id := f(e.Set.Id)
		if _, ok := m[id]; ok {
			return errAttemptToSetExistingId
		}
		m[id] = e.Set.Value
	case Tick:
		id := f(e.Tick)
		if _, ok := m[id]; !ok {
			return errAttemptToTickUnknownId
		}

		m[id] += 1
	case Merge:
		for id := range e.Merge {
			nid := f(id)
			if _, ok := m[nid]; ok {
				if m[nid] < e.Merge[id] {
					m[nid] = e.Merge[id]
				}
			} else {
				m[nid] = e.Merge[id]
			}
		}
	}
	return nil
}

// HistoryItem stores details of a state change due to the specified Event,
// and holds the updated clock after the Event has been applied.
type HistoryItem struct {
	HistoryId uint64
	Change    *Event
	Clock     Clock
}

// copy returns a deep copy of the instance
func (h *HistoryItem) copy() *HistoryItem {
	hi := &HistoryItem{
		HistoryId: h.HistoryId,
		Clock:     copyMap(h.Clock),
	}

	if h.Change != nil {
		hi.Change = h.Change.copy()
	}
	return hi
}

// copyWithKeyModification returns a deep copy of the instance,
// where keys in the Clock are adjusted by the function supplied.
func (h *HistoryItem) copyWithKeyModification(f func(string) (string, error)) (*HistoryItem, error) {
	m, err := copyMapWithKeyModification(h.Clock, f)
	if err != nil {
		return nil, err
	}

	hi := &HistoryItem{
		HistoryId: h.HistoryId,
		Clock:     m,
	}

	if h.Change != nil {
		hi.Change = h.Change.copy()
	}
	return hi, nil
}

func (h *HistoryItem) String() string {
	return fmt.Sprint(*h)
}
