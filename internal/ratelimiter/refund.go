package ratelimiter

import (
	"github.com/DarkIntaqt/cosmic-radiance/internal/request"
	"github.com/DarkIntaqt/cosmic-radiance/internal/schema"
)

type Refund struct {
	Syntax   *schema.Syntax
	Priority request.Priority
	KeyId    int
}

func (rl *RateLimiter) handleRefund(refund Refund) {
	if refund.Syntax == nil {
		return
	}
	queue := rl.queueManager.GetQueue(*refund.Syntax, refund.Priority)
	if queue != nil {
		queue.Refund(refund.KeyId)
	}
}
