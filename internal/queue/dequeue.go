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

		req, keyId := rb.dequeueDispatchable(now)

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

// dequeueDispatchable scans the queue for the first request that can be processed with an available API key
// It removes the request from the array, shifts preceding elements, and returns the request and the assigned KeyId.
func (rb *RingBuffer) dequeueDispatchable(now time.Time) (*request.Request, int) {
	if rb.count == 0 {
		return nil, -1
	}

	limits := rb.Limits

	anyAvailable := false
	for k := 0; k < len(*limits); k++ {
		if (*limits)[k].CanAllow(now, rb.Priority) {
			anyAvailable = true
			break
		}
	}
	if !anyAvailable {
		return nil, -1
	}

	for i := int64(0); i < rb.count; i++ {
		idx := (rb.head + i) % rb.size
		req := rb.entries[idx]

		if req == nil {
			continue
		}

		// Skip expired requests
		if now.UnixMilli() >= req.Expire {
			continue
		}

		availableKeyId := -1

		if req.TokenIndex != nil {
			// Specific token requested
			if *req.TokenIndex < len(*limits) && (*limits)[*req.TokenIndex].TryAllow(now, rb.Priority) {
				availableKeyId = *req.TokenIndex
			}
		} else {
			// Any token
			for k := 0; k < len(*limits); k++ {
				if (*limits)[k].TryAllow(now, rb.Priority) {
					availableKeyId = k
					break
				}
			}
		}

		if availableKeyId >= 0 {
			// We can dispatch this request.
			// Shift all elements before this one forward by 1 to fill the hole.
			if i > 0 {
				for j := i; j > 0; j-- {
					currIdx := (rb.head + j) % rb.size
					prevIdx := (rb.head + j - 1 + rb.size) % rb.size
					rb.entries[currIdx] = rb.entries[prevIdx]
				}
			}

			// Empty spot is now at the head, dequeue it normally
			rb.entries[rb.head] = nil
			rb.head = (rb.head + 1) % rb.size
			rb.count--

			if rb.count == 0 {
				rb.lastUpdated = time.Now()
			}

			return req, availableKeyId
		}
	}

	return nil, -1
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
