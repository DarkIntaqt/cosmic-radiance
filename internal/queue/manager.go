package queue

import (
	"log"
	"time"

	"github.com/DarkIntaqt/cosmic-radiance/configs"
	"github.com/DarkIntaqt/cosmic-radiance/internal/request"
	"github.com/DarkIntaqt/cosmic-radiance/internal/resource"
	"github.com/DarkIntaqt/cosmic-radiance/internal/schema"
	"github.com/DarkIntaqt/cosmic-radiance/ratelimiter/options"
)

type QueueManager struct {
	Queues              map[string]*RingBuffer
	PriorityQueues      map[string]*RingBuffer
	RateLimitGroups     map[string]*resource.RateLimitGroupSlice // per ID, holds several api keys
	RateLimitCategories []map[string]*resource.RateLimitCategory // for each api key, holds either platform or ID
	opts                *options.RateLimiterOptions
}

func NewQueueManager(opts *options.RateLimiterOptions) *QueueManager {

	manager := &QueueManager{
		Queues:              make(map[string]*RingBuffer),
		PriorityQueues:      make(map[string]*RingBuffer),
		RateLimitGroups:     make(map[string]*resource.RateLimitGroupSlice),
		RateLimitCategories: make([]map[string]*resource.RateLimitCategory, len(opts.ApiKeys)),
		opts:                opts,
	}

	// Init the maps
	for i := 0; i < len(opts.ApiKeys); i++ {
		manager.RateLimitCategories[i] = make(map[string]*resource.RateLimitCategory)
	}

	return manager
}

/*
Enqueues a request into the appropriate queue based on its priority.
Creates queues with rate limits if they don't exist yet.
Returns a timestamp when the request can be retried if the queue is full.
*/
func (qm *QueueManager) EnqueueRequest(req *request.Request, priority request.Priority, syntax *schema.Syntax) *time.Time {
	// Set the current queue
	queue := qm.getQueues(priority)

	// Check if the queue already exists
	if _, exists := queue[syntax.Id]; !exists {

		// Create if the rate limit group doesn't already exists
		if groups, exists := qm.RateLimitGroups[syntax.Id]; !exists || len(*groups) == 0 {

			now := time.Now()

			// Create the rate limit group and categories and fill it with placeholder limits
			// A group will only exists if a category also already exists
			rateLimitGroupSlice := make(resource.RateLimitGroupSlice, len(qm.opts.ApiKeys))
			qm.RateLimitGroups[syntax.Id] = &rateLimitGroupSlice
			created := 0

			for i := 0; i < len(qm.opts.ApiKeys); i++ {
				if _, Ok := qm.RateLimitCategories[i][syntax.Id]; !Ok {
					created++
					qm.RateLimitCategories[i][syntax.Id] = &resource.RateLimitCategory{
						LockedUntil: now,
						RateLimits: []*resource.RateLimit{{
							Window:     time.Duration(5 * time.Second),
							Limit:      5,
							Current:    0,
							LastRefill: now,
						}},
						AdditionalWindowSize: &qm.opts.AdditionalWindowSize,
						Timeout:              &qm.opts.Timeout,
					}
				}

				// Check if there are no limits for the platform already
				if _, Ok := qm.RateLimitCategories[i][syntax.Platform]; !Ok {
					created++
					qm.RateLimitCategories[i][syntax.Platform] = (&resource.RateLimitCategory{
						LockedUntil: now,
						RateLimits: []*resource.RateLimit{{
							Window:     time.Duration(5 * time.Second),
							Limit:      5,
							Current:    0,
							LastRefill: now,
						}},
						AdditionalWindowSize: &qm.opts.AdditionalWindowSize,
						Timeout:              &qm.opts.Timeout,
					})
				}

				platformLimits := qm.RateLimitCategories[i][syntax.Platform]
				methodLimits := qm.RateLimitCategories[i][syntax.Id]

				(*qm.RateLimitGroups[syntax.Id])[i] = &resource.RateLimitGroup{
					KeyId: i,
					// Instantly trigger an update by setting lastUpdated to the past
					LastUpdated:    now.Add(-1 * (configs.RATELIMIT_UPDATE_INTERVAL + 1*time.Second)),
					PlatformLimits: platformLimits,
					MethodLimits:   methodLimits,
					// Set peak capacity to something that smaller...
					PeakCapacity: int64(50 * qm.opts.Timeout.Seconds() / float64(i+1)),
				}
			}
		}

		groups := qm.RateLimitGroups[syntax.Id]
		queue[syntax.Id] = newRingBuffer(groups, priority, qm.opts.PriorityQueueSize)
		if priority == request.HighPriority {
			log.Printf("Queue #P-%s created for %s/%s with size of %d\n", syntax.Id, syntax.Platform, syntax.Endpoint, queue[syntax.Id].size)
		} else {
			log.Printf("Queue #%s created for %s/%s with size of %d\n", syntax.Id, syntax.Platform, syntax.Endpoint, queue[syntax.Id].size)
		}
	}

	// Enqueue the request into the right queue
	return queue[syntax.Id].Enqueue(req)
}

func (qm *QueueManager) getQueues(priority request.Priority) map[string]*RingBuffer {
	// Set the current queue
	queue := qm.Queues
	if priority == request.HighPriority {
		queue = qm.PriorityQueues
	}

	return queue
}

func (qm *QueueManager) GetQueue(syntax schema.Syntax, priority request.Priority) *RingBuffer {
	return qm.getQueues(priority)[syntax.Id]
}

func (qm *QueueManager) Drain() {
	for _, queue := range qm.Queues {
		queue.drain()
	}
	for _, queue := range qm.PriorityQueues {
		queue.drain()
	}
}

func (qm *QueueManager) AdjustQueueSize() {
	now := time.Now()
	for key, queue := range qm.getQueues(request.NormalPriority) {
		peakCapacity := queue.GetPeakCapacity()
		if peakCapacity != queue.size && queue.Count() < peakCapacity {
			newQueue := newRingBuffer(queue.Limits, queue.Priority, qm.opts.PriorityQueueSize)

			// put all entries from the old queue in the new queue
			for queue.Count() > 0 {
				newQueue.Enqueue(queue.Dequeue(now))
			}

			log.Printf("Queue #%s adjusted size from %d to %d\n", key, queue.size, newQueue.size)
			go queue.drain()
			qm.Queues[key] = newQueue
		}
	}

	for key, queue := range qm.getQueues(request.HighPriority) {
		// Priority queues only get a fraction an original queues size to prevent overflows and priority spamming
		peakCapacity := int64(float32(queue.GetPeakCapacity())*qm.opts.PriorityQueueSize) + 1
		if peakCapacity != queue.size && queue.Count() < peakCapacity {
			newQueue := newRingBuffer(queue.Limits, queue.Priority, qm.opts.PriorityQueueSize)

			// put all entries from the old queue in the new queue
			for queue.Count() > 0 {
				newQueue.Enqueue(queue.Dequeue(now))
			}

			log.Printf("Queue #P-%s adjusted size from %d to %d\n", key, queue.size, newQueue.size)
			go queue.drain()
			qm.PriorityQueues[key] = newQueue
		}
	}
}

func (qm *QueueManager) CleanUp() {
	now := time.Now()

	for key, queue := range qm.getQueues(request.NormalPriority) {
		if queue.Count() == 0 && queue.lastUpdated.Before(now.Add(-configs.QUEUE_INACTIVITY)) {
			// If the queue is empty and hasn't been updated in a while, we can remove it
			log.Printf("Queue #%s removed due to inactivity\n", key)
			qm.Queues[key].drain()
			delete(qm.Queues, key)
		}
	}
	for key, queue := range qm.getQueues(request.HighPriority) {
		if queue.Count() == 0 && queue.lastUpdated.Before(now.Add(-configs.QUEUE_INACTIVITY)) {
			// If the queue is empty and hasn't been updated in a while, we can remove it
			log.Printf("Queue #P-%s removed due to inactivity\n", key)
			qm.PriorityQueues[key].drain()
			delete(qm.PriorityQueues, key)
		}
	}

}
