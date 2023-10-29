[![Go Doc](https://pkg.go.dev/badge/github.com/gford1000-go/vclock.svg)](https://pkg.go.dev/github.com/gford1000-go/vclock)
[![Go Report Card](https://goreportcard.com/badge/github.com/gford1000-go/vclock)](https://goreportcard.com/report/github.com/gford1000-go/vclock)

vclock
======

vclock implements a vector clock, which can be used concurrently across goroutines.

The `VClock` instance can either maintain only the current state of the underlying `Clock`, or it 
can retain all the Events that create the history of change to the original `Clock`, in the 
order they are received.  This history is available as a slice of `*HistoryItem`.

The `VClock` obeys the state of the parent `context` that is passed into either `New` function, ensuring
that all resources are released correctly.  Should the parent `context` end, then all subsequent calls
to the `VClock` instance will return an `error`.

Vector clocks can be compared, and have four outcomes:
* They are equal; i.e. each identifier in the `Clock`s being compared have identical values
* One is the ancestor of the other.  Both clocks will include the identifiers within the ancestoral clock, with the ancestor having at least one identifier value that is smaller than in the other clock
* One is the descendant of the other.  This is essentially the reverse of ancestor.
* They are causally concurrent; i.e. there is no clear history linking them together.

Vector clocks can have identifiers that are arbitrarily long.  To keep the size of the `Clock` small, the `New` functions
include the argument `shortener` which is an interface of type `IdentifierShortener`.  If provided, then the Vector clock
will apply the functions from this interface to shorten the identifiers during updates, and recover the identifiers when
the `Clock` is returned externally.

There are examples of specific use cases within `example_test.go`, but general use looks as follows:


Usage:

```go
func main() {
    c, _ := New(context.Background(), Clock{"x":0, "y":0}, nil)
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
