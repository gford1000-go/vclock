[![Go Doc](https://pkg.go.dev/badge/github.com/gford1000-go/vclock.svg)](https://pkg.go.dev/github.com/gford1000-go/vclock)
[![Go Report Card](https://goreportcard.com/badge/github.com/gford1000-go/vclock)](https://goreportcard.com/report/github.com/gford1000-go/vclock)

vclock
=========

vclock implements a basic vector clock, which can be used concurrently across goroutines.

Usage:

```go
func main() {
    c, _ := New(map[string]uint64{"x":0, "y":0})
    defer c.Close()

    var wg sync.WaitGroup
    for i := 0; i < 10; i++ {
        wg.Add(1)
        go func(i int) {
            defer wg.Done()
            if i % 2 == 0 {
                c.Tick("x")
            } else {
                c.Tick("y")
            }
        }(i)        
    }

    wg.Wait()

    b, _ := c.Bytes()
}

```
