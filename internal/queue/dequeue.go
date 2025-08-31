package queue

import (
	"time"

	"github.com/DarkIntaqt/cosmic-radiance/internal/request"
)

/*
INTERNAL:
Dequeues a request from the RingBuffer.
*/
func (rb *RingBuffer) dequeue() *request.Request {
	if rb.count == 0 {
		return nil
	}

	req := rb.entries[rb.head]

	// Set reference to nil. This should hopefully erase memory if object is not referenced elsewhere
	rb.entries[rb.head] = nil

	rb.head = (rb.head + 1) % rb.size
	rb.count--

	// Set last updated to check for potential drops
	if rb.count == 0 {
		rb.lastUpdated = time.Now()
	}

	return req
}

/*
Dequeues the next available request from the RingBuffer that is not expired.
*/
func (rb *RingBuffer) Dequeue(now time.Time) *request.Request {
	return rb.purgeAndDequeue(now)
}

func (rb *RingBuffer) Process(max int) {
	now := time.Now()

	// Purge queues to reduce queue size
	rb.purge(now)

	for i := 0; i < max && rb.count > 0; i++ {

		keyId := rb.canDequeue(now)

		if keyId < 0 {
			break
		}

		req := rb.Dequeue(now)

		if req == nil {
			break
		}

		// Giving the request the corresponding key id
		req.Response <- &request.ResponseChannel{
			KeyId:  keyId,
			Update: rb.needsUpdate(keyId, now),
		}
	}
}

func (rb *RingBuffer) canDequeue(now time.Time) int {
	limits := rb.Limits

	// Cycle through the key and check for one that allows the request
	for i := 0; i < len(*limits); i++ {
		if (*limits)[i].TryAllow(now, rb.Priority) {
			return i
		}
	}

	return -1

}

func (rb *RingBuffer) needsUpdate(keyId int, now time.Time) bool {
	limits := rb.Limits

	// Key validation can be skipped, IT HAS TO EXISTS
	// if keyId < 0 || keyId >= len(*limits) {
	// 	return false
	// }

	// Check if the key's rate limit needs an update
	return (*limits)[keyId].NeedsUpdate(now)
}
