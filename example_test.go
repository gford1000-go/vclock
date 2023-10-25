package vclock

import (
	"fmt"
	"sync"
)

func ExampleNew() {
	c, _ := New(map[string]uint64{"x": 0, "y": 0})
	defer c.Close()

	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			if i%2 == 0 {
				c.Tick("x")
			} else {
				c.Tick("y")
			}
		}(i)
	}

	wg.Wait()

	m, _ := c.GetMap()
	fmt.Println(m)
	// Output: map[x:5 y:5]
}

func ExampleNewShowingTick() {

	// This example illustrates an implementation of the vector clock
	// example from https://en.wikipedia.org/wiki/Vector_clock

	newF := func(m map[string]uint64) *VClock { vc, _ := New(m); return vc }

	// process emulates an autonomous process with a globally unique identifier
	type process struct {
		id string
		vc *VClock
	}

	// Transmit is the sending and receipt of an event between processes
	transfer := func(from, to *process) {

		// Emulates sending a vector clock as part of an event moving between processes
		send := func(from *process) []byte {
			from.vc.Tick(from.id)
			b, _ := from.vc.Bytes()
			return b
		}

		// Emulates receiving a vector clock as part of an event moving between processes
		recv := func(to *process, b []byte) {
			to.vc.Tick(to.id)
			vc, _ := FromBytes(b)
			to.vc.Merge(vc)
		}

		// Perform the transfer
		recv(to, send(from))
	}

	// The three processes of the example
	var A *process = &process{id: "a", vc: newF(map[string]uint64{"a": 0})}
	var B *process = &process{id: "b", vc: newF(map[string]uint64{"b": 0})}
	var C *process = &process{id: "c", vc: newF(map[string]uint64{"c": 0})}

	// This is the event sequence of the example
	transfer(C, B)
	transfer(B, A)
	transfer(B, C)
	transfer(A, B)
	transfer(C, A)
	transfer(B, C)
	transfer(C, A)

	a, _ := A.vc.GetMap()
	fmt.Println(a)
	// Output: map[a:4 b:5 c:5]
}
