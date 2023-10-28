package vclock

import "testing"

func BenchmarkNew(b *testing.B) {

	for i := 0; i < b.N; i++ {
		c, _ := New(Clock{"a": 0}, nil)
		defer c.Close()
	}
}

func BenchmarkTick(b *testing.B) {

	c, _ := New(Clock{"a": 0}, nil)
	defer c.Close()

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

	for i := 0; i < b.N; i++ {
		c, _ := New(Clock{"a": 0}, nil)
		defer c.Close()

		c.Merge(other)
	}
}

func BenchmarkMergeWithHistory(b *testing.B) {

	other, _ := NewWithHistory(Clock{"b": 1}, nil)
	defer other.Close()

	for i := 0; i < b.N; i++ {
		c, _ := New(Clock{"a": 0}, nil)
		defer c.Close()

		c.Merge(other)
	}
}

func BenchmarkBytes(b *testing.B) {

	c, _ := New(Clock{"a": 1}, nil)
	defer c.Close()

	for i := 0; i < b.N; i++ {
		c.Bytes()
	}
}

func BenchmarkBytesWithHistory(b *testing.B) {

	c, _ := NewWithHistory(Clock{"a": 1}, nil)
	defer c.Close()

	for i := 0; i < b.N; i++ {
		c.Bytes()
	}
}

func BenchmarkFromBytes(b *testing.B) {

	c, _ := New(Clock{"a": 1}, nil)
	defer c.Close()

	buf, _ := c.Bytes()

	for i := 0; i < b.N; i++ {
		n, _ := FromBytes(buf, nil)
		n.Close()
	}
}

func BenchmarkFromBytesWithHistory(b *testing.B) {

	c, _ := NewWithHistory(Clock{"a": 1}, nil)
	defer c.Close()

	buf, _ := c.Bytes()

	for i := 0; i < b.N; i++ {
		n, _ := FromBytes(buf, nil)
		n.Close()
	}
}
