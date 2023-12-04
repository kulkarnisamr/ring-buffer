package main

import (
	"testing"
)

func TestEnqueue(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		items    []*LogEntry
		wantSize int
	}{
		{
			name:     "enqueue one item",
			capacity: 5,
			items:    []*LogEntry{{}},
			wantSize: 1,
		},
		{
			name:     "enqueue two items",
			capacity: 5,
			items:    []*LogEntry{{}, {}},
			wantSize: 2,
		},
		{
			name:     "enqueue more items than capacity",
			capacity: 2,
			items:    []*LogEntry{{}, {}, {}},
			wantSize: 3, // if your buffer allows more items than its capacity
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &RingBuffer{
				buffer:   make([]*LogEntry, 0, tt.capacity),
				capacity: tt.capacity,
			}

			for _, item := range tt.items {
				b.enqueue(item)
			}

			if b.size != tt.wantSize {
				t.Errorf("RingBuffer.enqueue() = %v, want %v", b.size, tt.wantSize)
			}
		})
	}
}
func TestDequeue(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		items    []*LogEntry
		wantSize int
	}{
		{
			name:     "dequeue from empty buffer",
			capacity: 5,
			items:    []*LogEntry{},
			wantSize: 0,
		},
		{
			name:     "dequeue one item",
			capacity: 5,
			items:    []*LogEntry{{}, {}},
			wantSize: 1,
		},
		{
			name:     "dequeue two items",
			capacity: 5,
			items:    []*LogEntry{{}, {}, {}},
			wantSize: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &RingBuffer{
				buffer:   make([]*LogEntry, 0, tt.capacity),
				capacity: tt.capacity,
			}

			for _, item := range tt.items {
				b.enqueue(item)
			}

			//			fmt.Println(len(tt.items) - tt.wantSize)
			for i := 0; i < len(tt.items)-tt.wantSize; i++ {
				b.dequeue()
			}

			if b.size != tt.wantSize {
				t.Errorf("RingBuffer.dequeue() = %v, want %v", b.size, tt.wantSize)
			}
		})
	}
}
