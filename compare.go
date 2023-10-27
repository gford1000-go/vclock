package vclock

// condition constants define how to compare a vector clock against another,
// and may be ORed together when being provided to the Compare method.
type condition int

// Constants define compairison conditions between pairs of vector clocks
const (
	equal      condition = 1 << iota // Clocks are identical
	ancestor                         // One clock is a clear ancestor to the other
	descendant                       // One clock is a clear descendent to the other
	concurrent                       // Clocks are completely independent, or partial overlap
)

type respComp struct {
	other map[string]uint64
	cond  condition
}

// Compare takes another clock and determines if it is equal, an
// ancestor, descendant, or concurrent with the callees clock.
func compare(vc, other map[string]uint64, cond condition) bool {

	var otherIs condition
	// Preliminary qualification based on length
	if len(vc) > len(other) {
		if cond&(ancestor|concurrent) == 0 {
			return false
		}
		otherIs = ancestor
	} else if len(vc) < len(other) {
		if cond&(descendant|concurrent) == 0 {
			return false
		}
		otherIs = descendant
	} else {
		otherIs = equal
	}

	keys := sortedKeys(vc)
	otherKeys := sortedKeys(other)

	if cond&(equal|descendant) != 0 {
		// All of the identifiers in this clock must be present in the other
		for _, key := range keys {
			if _, ok := other[key]; !ok {
				return false
			}
		}
	}
	if cond&(equal|ancestor) != 0 {
		// All of the identifiers in this clock must be present in the other
		for _, key := range otherKeys {
			if _, ok := vc[key]; !ok {
				return false
			}
		}
	}

	// Compare matching items, using sortedKeys to provide
	// deterministic (sorted) comparison of lock identifiers
	for _, id := range otherKeys {
		if _, found := vc[id]; found {
			if other[id] > vc[id] {
				switch otherIs {
				case equal:
					if cond&descendant == 0 {
						return false
					}
					otherIs = descendant
				case ancestor:
					return cond&concurrent != 0
				}
			} else if other[id] < vc[id] {
				switch otherIs {
				case equal:
					if cond&ancestor == 0 {
						return false
					}
					otherIs = ancestor
				case descendant:
					return cond&concurrent != 0
				}
			}
		} else {
			if otherIs == equal {
				return cond&concurrent != 0
			} else if (len(other) - len(vc) - 1) < 0 {
				return cond&concurrent != 0
			}
		}
	}
	return cond&otherIs != 0
}
