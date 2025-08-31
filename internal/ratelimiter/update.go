package ratelimiter

import (
	"math"
	"net/http"
	"time"

	"github.com/DarkIntaqt/cosmic-radiance/internal/request"
	"github.com/DarkIntaqt/cosmic-radiance/internal/schema"
)

type LimitType int

const (
	PlatformLimit LimitType = iota
	MethodLimit
)

type Update struct {
	syntax     *schema.Syntax
	header     *http.Header
	keyId      int
	priority   request.Priority
	RetryAfter *time.Time
	LimitType  LimitType
}

func (rl *RateLimiter) updateRatelimits(syntax *schema.Syntax, response *http.Response, keyId int, priority request.Priority) {
	if syntax == nil || response == nil {
		return
	}

	var retryAfter *time.Time
	var limitType LimitType

	if response.StatusCode == http.StatusTooManyRequests {
		if ra := response.Header.Get("Retry-After"); ra != "" {
			if dur, err := time.ParseDuration(ra + "s"); err == nil {
				t := time.Now().Add(dur)
				retryAfter = &t

				if rt := response.Header.Get("X-Rate-Limit-Type"); rt != "" {
					switch rt {
					case "platform":
						limitType = PlatformLimit
					case "method":
						limitType = MethodLimit
					}
				}
			}
		}
	}

	rl.updateChannel <- Update{
		syntax:     syntax,
		keyId:      keyId,
		priority:   priority,
		header:     &response.Header,
		RetryAfter: retryAfter,
		LimitType:  limitType,
	}
}

func (rl *RateLimiter) handleUpdateRequest(update Update) {
	if update.syntax == nil || update.header == nil {
		return
	}

	// Update the rate limits in the queue manager
	manager := rl.queueManager

	// Set last updated to now, dunno if that even works
	queue := manager.GetQueue(*update.syntax, update.priority)
	limits := (*queue.Limits)[update.keyId]

	peakCapacity := math.Min(
		limits.PlatformLimits.Update((*update.header).Get("X-App-Rate-Limit"), (*update.header).Get("X-App-Rate-Limit-Count"), update.RetryAfter, update.LimitType == PlatformLimit),
		limits.MethodLimits.Update((*update.header).Get("X-Method-Rate-Limit"), (*update.header).Get("X-Method-Rate-Limit-Count"), update.RetryAfter, update.LimitType == MethodLimit),
	)

	if peakCapacity > 0 {
		limits.PeakCapacity = int64((peakCapacity + 1) * 1.05)
		limits.LastUpdated = time.Now()
	}
}
