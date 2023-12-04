package main

import (
	"testing"
)

// BenchmarkWrite tests the performance of the Write operation.
func BenchmarkWrite(b *testing.B) {
	rb := NewRingBuffer(1024)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rb.enqueue(&LogEntry{
			entryIndex: int64(i),
			data:       []byte("hello world"),
		})
	}
}

// BenchmarkRead tests the performance of the Read operation.
func BenchmarkRead(b *testing.B) {
	rb := NewRingBuffer(1024)

	for i := 0; i < b.N; i++ {
		rb.enqueue(&LogEntry{
			entryIndex: int64(i),
			data:       []byte("hello world"),
		})
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = rb.dequeue()
	}
}
