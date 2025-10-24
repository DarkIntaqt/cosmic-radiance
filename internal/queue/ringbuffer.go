package queue

import (
	"time"

	"github.com/DarkIntaqt/cosmic-radiance/configs"
	"github.com/DarkIntaqt/cosmic-radiance/internal/request"
	"github.com/DarkIntaqt/cosmic-radiance/internal/resource"
)

// This ring buffer should be atomic. It can only be read by the main thread.
// There is no need to lock the design
// Dequeued *request.Request pointers should be set to nil in order to GC them

type RingBuffer struct {
	entries []*request.Request
	size    int64
	// Pointing to the boundaries of the queue
	head, tail int64
	// Size of the queue
	count int64
	// Current unused
	lastUpdated time.Time
	Priority    request.Priority

	// List of rate limits. One group per API key
	Limits *resource.RateLimitGroupSlice
}

func newRingBuffer(limits *resource.RateLimitGroupSlice, priority request.Priority) *RingBuffer {
	buffer := &RingBuffer{
		head:        0,
		tail:        0,
		count:       0,
		lastUpdated: time.Now(),
		Priority:    priority,
		Limits:      limits,
	}

	size := buffer.GetPeakCapacity()

	if priority == request.HighPriority {
		size = int64(float32(buffer.GetPeakCapacity())*configs.PriorityQueueSize) + 1
	}

	buffer.entries = make([]*request.Request, size)
	buffer.size = size

	return buffer
}

func (rb *RingBuffer) Count() int64 {
	return rb.count
}

func (rb *RingBuffer) Refund(keyId int, timestamp time.Time) {
	if keyId < 0 || keyId >= len(*rb.Limits) {
		return
	}
	(*rb.Limits)[keyId].Refund(timestamp)
}
