package queue

import (
	"time"

	"github.com/DarkIntaqt/cosmic-radiance/internal/request"
)

/*
INTERNAL:
peek into the next request
*/
func (rb *RingBuffer) peek() *request.Request {
	if rb.count == 0 {
		return nil
	}

	return rb.entries[rb.head]
}

/*
Peek returns the next valid request without removing it from the RingBuffer.
*/
func (rb *RingBuffer) Peek(now time.Time) *request.Request {
	return rb.purgeAndPeek(now)
}
