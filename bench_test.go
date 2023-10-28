package vclock

import (
	"bytes"
	"encoding/gob"
	"fmt"
	"testing"

	"golang.org/x/exp/rand"
)

func BenchmarkNewWithDefer(b *testing.B) {

	for i := 0; i < b.N; i++ {
		c, _ := New(Clock{"a": 0}, nil)
		defer c.Close()
	}
}

func BenchmarkTick(b *testing.B) {

	c, _ := New(Clock{"a": 0}, nil)
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

	c, _ := NewWithHistory(Clock{"a": 0}, nil)
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

	other, _ := New(Clock{"b": 1}, nil)
	defer other.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := New(Clock{"a": 0}, nil)
		defer c.Close()

		c.Merge(other)
	}
}

func BenchmarkMergeLarge(b *testing.B) {

	c := Clock{}
	for i := 0; i < 1024; i++ {
		c[fmt.Sprint(i)] = rand.Uint64()
	}

	other, _ := New(c, nil)
	defer other.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := New(Clock{"a": 0}, nil)
		defer c.Close()

		c.Merge(other)
	}
}

func BenchmarkMergeWithHistory(b *testing.B) {

	other, _ := NewWithHistory(Clock{"b": 1}, nil)
	defer other.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c, _ := New(Clock{"a": 0}, nil)
		defer c.Close()

		c.Merge(other)
	}
}

func BenchmarkBytes(b *testing.B) {

	c, _ := New(Clock{"a": 1}, nil)
	defer c.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Bytes()
	}
}

func BenchmarkBytesLarge(b *testing.B) {

	c := Clock{}
	for i := 0; i < 1024; i++ {
		c[fmt.Sprint(i)] = rand.Uint64()
	}

	vc, _ := New(c, nil)
	defer vc.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		vc.Bytes()
	}
}

func BenchmarkBytesWithHistory(b *testing.B) {

	c, _ := NewWithHistory(Clock{"a": 1}, nil)
	defer c.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Bytes()
	}
}

func BenchmarkFromBytes(b *testing.B) {

	c, _ := New(Clock{"a": 1}, nil)
	defer c.Close()

	buf, _ := c.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n, _ := FromBytes(buf, nil)
		n.Close()
	}
}

func BenchmarkPrepGOBDecoder(b *testing.B) {

	c, _ := New(Clock{"a": 1}, nil)
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

	c, _ := New(Clock{"a": 1}, nil)
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

	c, _ := NewWithHistory(Clock{"a": 1}, nil)
	defer c.Close()

	buf, _ := c.Bytes()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		n, _ := FromBytes(buf, nil)
		n.Close()
	}
}

func BenchmarkNewOnly(b *testing.B) {

	for i := 0; i < b.N; i++ {
		New(Clock{"a": 0}, nil)
	}
}
