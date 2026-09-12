package ratelimiter

import (
	"log/slog"

	"github.com/DarkIntaqt/cosmic-radiance/internal/request"
	"github.com/DarkIntaqt/cosmic-radiance/internal/schema"
)

type IncomingRequest struct {
	Request  *request.Request
	Syntax   *schema.Syntax
	Priority request.Priority
}

/*
INTERNAL:
Validate the incoming request and the enqueue it
*/
func (rl *RateLimiter) handleIncomingRequest(req IncomingRequest) {
	if req.Request == nil || req.Syntax == nil {
		return
	}

	// Try to enqueue request. If unsuccessful, return retry-after
	if time := rl.queueManager.EnqueueRequest(req.Request, req.Priority, req.Syntax); time != nil {
		slog.Debug("Request rejected due to full queue", "queueId", req.Syntax.Id, "platform", req.Syntax.Platform, "endpoint", req.Syntax.Endpoint, "retryAfter", time)
		req.Request.FailedResponse(time)
	}
}
