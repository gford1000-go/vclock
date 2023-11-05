package vclock

// import (
// 	"errors"
// 	"sync"
// 	"testing"
// 	"time"
// )

// func TestNewChannel(t *testing.T) {

// 	c1 := NewChannel[int]()
// 	defer c1.Close()

// 	c2 := NewChannel[bool]()
// 	defer c2.Close()

// 	for i := 0; i < 10000; i++ {

// 		go func(n int) {
// 			i, _ := c1.Recv()
// 			c2.Send(i == n)
// 		}(i)

// 		c1.Send(i)
// 		res, err := c2.Recv()

// 		if err != nil {
// 			t.Fatalf("unexpected error: %v", err)
// 		}
// 		if !res {
// 			t.Fatalf("unexpected receive on: %v", i)
// 		}
// 	}
// }

// func TestSendOnClosedChannel(t *testing.T) {

// 	c := NewChannel[int]()
// 	c.Close()

// 	err := c.Send(9)
// 	if err == nil {
// 		t.Fatal("should have received error")
// 	}
// 	if !errors.Is(err, ErrChannelClosed) {
// 		t.Fatalf("unexpected error: got (%v), expected (%v)", err, ErrChannelClosed)
// 	}

// }

// func TestRecvOnClosedChannel(t *testing.T) {

// 	c := NewChannel[int]()
// 	c.Close()

// 	_, err := c.Recv()
// 	if err == nil {
// 		t.Fatal("should have received error")
// 	}
// 	if !errors.Is(err, ErrChannelClosed) {
// 		t.Fatalf("unexpected error: got (%v), expected (%v)", err, ErrChannelClosed)
// 	}

// }

// func TestRawOnClosedChannel(t *testing.T) {

// 	c := NewChannel[int]()
// 	c.Close()

// 	ch := c.RawChan()
// 	if ch != nil {
// 		t.Fatal("should have received nil")
// 	}
// }

// func TestBufferedChan(t *testing.T) {

// 	c := NewChannel[int](10)
// 	defer c.Close()

// 	exit := NewChannel[bool]()
// 	defer exit.Close()

// 	exited := NewChannel[bool]()
// 	defer exited.Close()

// 	total := 0

// 	go func() {
// 		for {
// 			select {
// 			case i := <-c.RawChan():
// 				total += i
// 			case <-exit.RawChan():
// 				{
// 					exited.Send(true)
// 					return
// 				}
// 			}
// 		}
// 	}()

// 	N := 1000

// 	for i := 0; i < N; i++ {
// 		c.Send(i)
// 	}

// 	time.Sleep(1 * time.Millisecond)

// 	exit.Send(true)
// 	exited.Recv()

// 	NTot := N * (N - 1) / 2
// 	if total != NTot {
// 		t.Fatalf("expected %v, got %v", NTot, total)
// 	}
// }

// func TestBufferedChan2(t *testing.T) {

// 	c := NewChannelWithTimeout[int](1*time.Millisecond, 10)
// 	defer c.Close()

// 	var wg sync.WaitGroup
// 	wg.Add(1)

// 	total := 0

// 	go func() {
// 		defer wg.Done()
// 		for {
// 			i, err := c.Recv()
// 			if err != nil {
// 				return
// 			}
// 			total += i
// 		}
// 	}()

// 	N := 1000

// 	for i := 0; i < N; i++ {
// 		c.Send(i)
// 	}

// 	wg.Wait()

// 	NTot := N * (N - 1) / 2
// 	if total != NTot {
// 		t.Fatalf("expected %v, got %v", NTot, total)
// 	}
// }
