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
	var A *process = &process{id: "a", vc: newF(Clock{"a": 0})}
	var B *process = &process{id: "b", vc: newF(Clock{"b": 0})}
	var C *process = &process{id: "c", vc: newF(Clock{"c": 0})}

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

func ExampleGetHistory() {
	c, _ := NewWithHistory(Clock{"x": 0, "y": 0})
	defer c.Close()

	c.Tick("x")
	c.Tick("x")
	c.Tick("y")
	c.Tick("x")

	history, _ := c.GetHistory()
	fmt.Println(history)
	// Output: [map[x:0 y:0] map[x:1 y:0] map[x:2 y:0] map[x:2 y:1] map[x:3 y:1]]
}

func ExampleGetFullHistory() {
	c1, _ := NewWithHistory(Clock{"x": 0, "y": 0})
	defer c1.Close()

	c1.Tick("x")
	c1.Tick("x")
	c1.Tick("y")
	c1.Tick("x")

	// Show all possible Event types by merging another clock
	c2, _ := New(Clock{"z": 7})
	defer c2.Close()
	c1.Merge(c2)

	// This is quite confusing when printed, but illustrates the availability of detailed history information
	history, _ := c1.GetFullHistory()

	fmt.Println(history)
	// Output: [{0 <nil> map[x:0 y:0]} {1 {Tick <nil> x map[]} map[x:1 y:0]} {2 {Tick <nil> x map[]} map[x:2 y:0]} {3 {Tick <nil> y map[]} map[x:2 y:1]} {4 {Tick <nil> x map[]} map[x:3 y:1]} {5 {Merge <nil>  map[z:7]} map[x:3 y:1 z:7]}]
}

func ExamplePrune() {
	c, _ := NewWithHistory(Clock{"x": 0, "y": 0})
	defer c.Close()

	c.Tick("x")
	c.Tick("x")
	c.Tick("y")
	c.Tick("x")

	c.Prune()
	history, _ := c.GetHistory()
	fmt.Println(history)
	// Output: [map[x:3 y:1]]
}
