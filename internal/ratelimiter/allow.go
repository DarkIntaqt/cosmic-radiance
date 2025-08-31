package ratelimiter

import (
	"github.com/DarkIntaqt/cosmic-radiance/internal/queue"
	"github.com/DarkIntaqt/cosmic-radiance/internal/request"
)

func (rm *RateLimiter) processQueues(queues map[string]*queue.RingBuffer) {

	batchSize := 5
	// Allow priority requests in a ratio of 5/1

	for _, queue := range queues {
		if queue.Priority == request.HighPriority {
			batchSize = 25
		}
		queue.Process(batchSize)
	}

}
