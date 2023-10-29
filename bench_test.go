package vclock

import (
	"bytes"
	"context"
	"encoding/gob"
	"fmt"
	"testing"

	"golang.org/x/exp/rand"
)

func BenchmarkNewWithDefer(b *testing.B) {

	ctx := context.Background()

	for i := 0; i < b.N; i++ {
		c, _ := New(ctx, Clock{"a": 0}, nil)
		defer c.Close()
	}
}

func BenchmarkTick(b *testing.B) {

	ctx := context.Background()

	c, _ := New(ctx, Clock{"a": 0}, nil)
	defer c.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Tick("a")
	}

	if v, ok := c.Get("a"); !ok || v != uint64(b.N) {
		b.Fatalf("Clock has wrong value: expected %v, got %v\n", b.N, v)
	}
}

func BenchmarkTickWithHistory(b *testing.B) {

	ctx := context.Background()

	c, _ := NewWithHistory(ctx, Clock{"a": 0}, nil)
	defer c.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Tick("a")
	}

	if v, ok := c.Get("a"); !ok || v != uint64(b.N) {
		b.Fatalf("Clock has wrong value: expected %v, got %v\n", b.N, v)
	}
}

func BenchmarkMerge(b *testing.B) {
	ctx := context.Background()

	other, _ := New(ctx, Clock{"b": 1}, nil)
	defer other.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := New(ctx, Clock{"a": 0}, nil)
		defer c.Close()

		c.Merge(other)
	}
}

func BenchmarkMergeLarge(b *testing.B) {

	ctx := context.Background()

	c := Clock{}
	for i := 0; i < 1024; i++ {
		c[fmt.Sprint(i)] = rand.Uint64()
	}

	other, _ := New(ctx, c, nil)
	defer other.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := New(ctx, Clock{"a": 0}, nil)
		defer c.Close()

		c.Merge(other)
	}
}

func BenchmarkMergeWithHistory(b *testing.B) {

	ctx := context.Background()

	other, _ := NewWithHistory(ctx, Clock{"b": 1}, nil)
	defer other.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := New(ctx, Clock{"a": 0}, nil)
		defer c.Close()

		c.Merge(other)
	}
}

func BenchmarkBytes(b *testing.B) {

	ctx := context.Background()

	c, _ := New(ctx, Clock{"a": 1}, nil)
	defer c.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Bytes()
	}
}

func BenchmarkBytesLarge(b *testing.B) {

	ctx := context.Background()

	c := Clock{}
	for i := 0; i < 1024; i++ {
		c[fmt.Sprint(i)] = rand.Uint64()
	}

	vc, _ := New(ctx, c, nil)
	defer vc.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vc.Bytes()
	}
}

func BenchmarkBytesWithHistory(b *testing.B) {

	ctx := context.Background()

	c, _ := NewWithHistory(ctx, Clock{"a": 1}, nil)
	defer c.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Bytes()
	}
}

func BenchmarkFromBytes(b *testing.B) {

	ctx := context.Background()

	c, _ := New(ctx, Clock{"a": 1}, nil)
	defer c.Close()

	buf, _ := c.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n, _ := FromBytes(ctx, buf, nil)
		n.Close()
	}
}

func BenchmarkPrepGOBDecoder(b *testing.B) {

	ctx := context.Background()

	c, _ := New(ctx, Clock{"a": 1}, nil)
	defer c.Close()

	buf, _ := c.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		b := new(bytes.Buffer)
		b.Write(buf)
		gob.NewDecoder(b)
	}

}

func BenchmarkDeserialiseClock(b *testing.B) {

	ctx := context.Background()

	c, _ := New(ctx, Clock{"a": 1}, nil)
	defer c.Close()

	buf, _ := c.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {

		b := new(bytes.Buffer)
		b.Write(buf)
		dec := gob.NewDecoder(b)

		m := Clock{}
		dec.Decode(&m)
	}

}

func BenchmarkFromBytesWithHistory(b *testing.B) {

	ctx := context.Background()

	c, _ := NewWithHistory(ctx, Clock{"a": 1}, nil)
	defer c.Close()

	buf, _ := c.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n, _ := FromBytes(ctx, buf, nil)
		n.Close()
	}
}

func BenchmarkNewOnly(b *testing.B) {

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		New(ctx, Clock{"a": 0}, nil)
	}
}
