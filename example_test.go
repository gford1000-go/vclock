package vclock

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

func ExampleNew() {
	ctx := context.Background()

	c, _ := New(ctx, Clock{"x": 0, "y": 0}, "")
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

	m, _ := c.GetClock()
	fmt.Println(m)
	// Output: map[x:5 y:5]
}

func ExampleNewShowingTick() {
	ctx := context.Background()

	// This example illustrates an implementation of the vector clock
	// example from https://en.wikipedia.org/wiki/Vector_clock

	newF := func(m Clock) *VClock { vc, _ := New(ctx, m, ""); return vc }

	// process emulates an autonomous process with a globally unique identifier
	type process struct {
		id string
		vc *VClock
	}

	// The three processes of the example
	var A *process = &process{id: "a", vc: newF(Clock{"a": 0})}
	var B *process = &process{id: "b", vc: newF(Clock{"b": 0})}
	var C *process = &process{id: "c", vc: newF(Clock{"c": 0})}

	// notification transfers the clock and then sequentially handles the next notifications
	type notification struct {
		from *process
		to   *process
		next []*notification
	}

	// This is the event sequence of the example
	n := &notification{
		from: C,
		to:   B,
		next: []*notification{
			{
				from: B,
				to:   A,
				next: []*notification{
					{
						from: A,
						to:   B,
						next: []*notification{
							{
								from: B,
								to:   C,
								next: []*notification{
									{
										from: C,
										to:   A,
									},
								},
							},
						},
					},
				},
			},
			{
				from: B,
				to:   C,
				next: []*notification{
					{
						from: C,
						to:   A,
					},
				},
			},
		},
	}

	var wg sync.WaitGroup

	// sleep emulates real-world transfer wait periods
	sleep := func() time.Duration { return time.Duration(rand.Intn(50)) * time.Millisecond }

	// simRecursive type allows our inline func to work
	type simRecursive func(*notification, simRecursive)

	// runSim ensures that the WaitGroup is incremented prior to
	// starting a goroutine with the f()
	runSim := func(n *notification, f simRecursive) {
		wg.Add(1)
		go f(n, f)
	}

	// simulator is a recursive processor of the supplied notification instance
	simulator := func(n *notification, f simRecursive) {
		defer wg.Done()

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
				vc, _ := FromBytes(ctx, b, "")
				to.vc.Merge(vc)
			}

			// Perform the transfer
			recv(to, send(from))
		}

		// Random pause
		d := sleep()
		if d > 0 {
			time.Sleep(d)
		}

		// Transfer the VClock
		transfer(n.from, n.to)

		// Process any next notifications
		for _, nn := range n.next {
			runSim(nn, f)
		}
	}

	// Start the simulation
	runSim(n, simulator)

	// Wait for all transfers to complete
	wg.Wait()

	// Retrieve the Clock and see what we have
	a, _ := A.vc.GetClock()

	fmt.Println(a)
	// Output: map[a:4 b:5 c:5]
}

func ExampleGetHistory() {
	ctx := context.Background()

	c, _ := NewWithHistory(ctx, Clock{"x": 0, "y": 0}, "")
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
	ctx := context.Background()

	c1, _ := NewWithHistory(ctx, Clock{"x": 0, "y": 0}, "")
	defer c1.Close()

	c1.Tick("x")
	c1.Tick("x")
	c1.Tick("y")
	c1.Tick("x")

	// Show all possible Event types by merging another clock
	c2, _ := New(ctx, Clock{"z": 7}, "")
	defer c2.Close()
	c1.Merge(c2)

	// This is quite confusing when printed, but illustrates the availability of detailed history information
	history, _ := c1.GetFullHistory()

	fmt.Println(history)
	// Output: [{0 <nil> map[x:0 y:0]} {1 {Tick <nil> x map[]} map[x:1 y:0]} {2 {Tick <nil> x map[]} map[x:2 y:0]} {3 {Tick <nil> y map[]} map[x:2 y:1]} {4 {Tick <nil> x map[]} map[x:3 y:1]} {5 {Merge <nil>  map[z:7]} map[x:3 y:1 z:7]}]
}

func ExamplePrune() {
	ctx := context.Background()

	c, _ := NewWithHistory(ctx, Clock{"x": 0, "y": 0}, "")
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

func ExampleContextCancelled() {
	ctx, cancel := context.WithCancel(context.Background())

	vc, _ := NewWithHistory(ctx, Clock{"x": 0, "y": 0}, "")
	defer vc.Close()

	// Underlying context is cancelled ...
	// Sleep to allow context goroutine to run
	cancel()
	time.Sleep(100 * time.Microsecond)

	// Vector clock will be closed
	_, err := vc.GetClock()

	fmt.Println(err)
	// Output: attempt to interact with closed clock
}
