package queue

import (
	"time"

	"github.com/DarkIntaqt/cosmic-radiance/internal/request"
)

/*
Enqueues a request into the RingBuffer.
Returns a timestamp when the unsuccessful in order to notify when to try again
*/
func (rb *RingBuffer) Enqueue(req *request.Request) *time.Time {

	// Check if the ring buffer is full and nothing is purgable
	// Get expires timestamp of oldest request, then it should be possible to queue new requests.
	// TODO: thats probably a bad idea
	now := time.Now()
	if rb.size == rb.count && rb.purge(now) == 0 {

		entry := rb.peek()

		// Technically it is impossible for the entry to be nil here, but just in case
		if entry != nil {
			tryAgain := time.UnixMilli(entry.Expire)
			return &tryAgain
		}

		// This should (as seen above) not happen, if it does, return in one second
		tryAgain := now.Add(1 * time.Second)
		return &tryAgain

	}

	// Enqueue new request at the tail, then progress tail
	rb.entries[rb.tail] = req
	rb.tail = (rb.tail + 1) % rb.size
	rb.count++

	// nil is returned if the enqueue was successful
	return nil

}
