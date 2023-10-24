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
