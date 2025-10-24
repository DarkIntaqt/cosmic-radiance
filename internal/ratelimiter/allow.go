package ratelimiter

import (
	"github.com/DarkIntaqt/cosmic-radiance/configs"
	"github.com/DarkIntaqt/cosmic-radiance/internal/queue"
	"github.com/DarkIntaqt/cosmic-radiance/internal/request"
)

func (rm *RateLimiter) processQueues(queues map[string]*queue.RingBuffer) {

	batchSize := configs.MAX_BATCH_SIZE_NORMAL

	for _, queue := range queues {
		if queue.Priority == request.HighPriority {
			batchSize = configs.MAX_BATCH_SIZE_PRIORITY
		}
		queue.Process(batchSize)
	}

}
