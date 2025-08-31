package queue

import (
	"time"

	"github.com/DarkIntaqt/cosmic-radiance/internal/request"
)

// NOT TO BE CONFUSED WITH DRAIN!

/*
INTERNAL:
purge removes all outdated entries by peeking into them, then return amount of purged entries
*/
func (rb *RingBuffer) purge(nowTime time.Time) int {
	now := nowTime.UnixMilli()
	count := 0

	for {
		req := rb.peek()
		if req == nil {
			return count
		}

		if now < req.Expire {
			return count
		}

		count++
		rb.dequeue()
	}
}

/*
INTERNAL:
purgeAndPeek removes all outdated entries and returns the next valid request.
*/
func (rb *RingBuffer) purgeAndPeek(nowTime time.Time) *request.Request {
	now := nowTime.UnixMilli()

	for {
		req := rb.peek()
		if req == nil {
			return nil
		}

		if now < req.Expire {
			return req
		}

		rb.dequeue()
	}

}

/*
INTERNAL:
purgeAndDequeue removes all outdated entries and dequeues the next valid request.
*/
func (rb *RingBuffer) purgeAndDequeue(nowTime time.Time) *request.Request {
	now := nowTime.UnixMilli()

	for {
		req := rb.dequeue()
		if req == nil {
			return nil
		}

		if now < req.Expire {
			return req
		}
	}
}
