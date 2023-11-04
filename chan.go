package vclock

import (
	"errors"
	"sync"
	"time"
)

func NewChannelWithTimeout[T any](d time.Duration, size ...int) *Channel[T] {
	return newChannel[T](d, size...)
}

func NewChannel[T any](size ...int) *Channel[T] {
	return newChannel[T](0, size...)
}

func newChannel[T any](d time.Duration, size ...int) *Channel[T] {
	var ch chan T
	if len(size) == 0 {
		ch = make(chan T)
	} else {
		ch = make(chan T, size[0])
	}
	return &Channel[T]{
		c: ch,
		d: d,
	}
}

type Channel[T any] struct {
	l sync.RWMutex
	c chan T
	d time.Duration
}

func (c *Channel[T]) Close() error {
	c.l.Lock()
	defer c.l.Unlock()

	close(c.c)
	c.c = nil
	return nil
}

var ErrChannelClosed = errors.New("chan is closed")
var ErrChannelTimeout = errors.New("chan timed out")

func (c *Channel[T]) Send(t T) error {
	c.l.RLock()
	defer c.l.RUnlock()

	if c.c == nil {
		return ErrChannelClosed
	}

	c.c <- t
	return nil
}

func (c *Channel[T]) Recv() (T, error) {
	c.l.RLock()
	defer c.l.RUnlock()

	var t T
	if c.c == nil {
		return t, ErrChannelClosed
	}

	if c.d == 0 {
		return <-c.c, nil
	}

	select {
	case v := <-c.c:
		return v, nil
	case <-time.After(c.d):
		return t, ErrChannelTimeout
	}
}

func (c *Channel[T]) RawChan() chan T {
	c.l.RLock()
	defer c.l.RUnlock()
	return c.c
}
