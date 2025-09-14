package resource

import (
	"math"
	"time"

	"github.com/DarkIntaqt/cosmic-radiance/configs"
	"github.com/DarkIntaqt/cosmic-radiance/internal/request"
)

type RateLimitGroupSlice = []*RateLimitGroup

type RateLimitGroup struct {
	PlatformLimits *RateLimitCategory
	MethodLimits   *RateLimitCategory
	KeyId          int
	LastUpdated    time.Time
	PeakCapacity   int64
	// TotalRequests  int64 // counter of total requests for analytics
}

/*
Returns whether the rate limit on that key needs to be refreshed
*/
func (rlg *RateLimitGroup) NeedsUpdate(now time.Time) bool {
	// Update limits every five minutes
	needsUpdate := rlg.LastUpdated.Before(now.Add(-configs.RATELIMIT_UPDATE_INTERVAL))

	if needsUpdate {
		rlg.LastUpdated = rlg.LastUpdated.Add(configs.UPDATE_GRATUITY)
	}

	return needsUpdate
}

/*
Tries to allow a request through the rate limit.
If the request is allowed, it consumes the available quota.
*/
func (rlg *RateLimitGroup) TryAllow(now time.Time, priority request.Priority) bool {
	if rlg.PlatformLimits.LockedUntil.After(now) || rlg.MethodLimits.LockedUntil.After(now) {
		return false
	}

	// Loop through all available limits
	for i := range rlg.PlatformLimits.RateLimits {
		rl := rlg.PlatformLimits.RateLimits[i]

		if rl.Current >= rl.Limit {
			return false
		}

		if priority == request.HighPriority {
			continue
		}

		// Compute how many requests should have been allowed by now ideally
		elapsed := now.Sub(rl.LastRefill)
		idealAllowed := int(math.Min(float64(elapsed)/float64(rl.Window)*float64(rl.Limit)+1, float64(rl.Limit)))

		if rl.Current > idealAllowed {
			return false
		}
	}

	for i := range rlg.MethodLimits.RateLimits {
		rl := rlg.MethodLimits.RateLimits[i]

		if rl.Current >= rl.Limit {
			return false
		}

		if priority == request.HighPriority {
			continue
		}

		// Compute how many requests should have been allowed by now ideally
		elapsed := now.Sub(rl.LastRefill)
		idealAllowed := int(math.Min(float64(elapsed)/float64(rl.Window)*float64(rl.Limit)+1, float64(rl.Limit)))

		if rl.Current > idealAllowed {
			return false
		}
	}

	// If all is set, consume all available limits
	for i := range rlg.PlatformLimits.RateLimits {
		rl := rlg.PlatformLimits.RateLimits[i]
		if rl.Current == 0 {
			rl.LastRefill = now
		}
		rl.Current++
	}

	for i := range rlg.MethodLimits.RateLimits {
		rl := rlg.MethodLimits.RateLimits[i]
		if rl.Current == 0 {
			rl.LastRefill = now
		}
		rl.Current++
	}

	// Increase total request amount for that endpoint
	// rlg.TotalRequests++

	return true
}

/*
Refunds a request.
Inverse of TryAllow.
*/
func (rlg *RateLimitGroup) Refund() {
	// Decrease all limits by one
	for i := range rlg.PlatformLimits.RateLimits {
		rl := rlg.PlatformLimits.RateLimits[i]
		if rl.Current > 0 {
			rl.Current--
		}
	}

	for i := range rlg.MethodLimits.RateLimits {
		rl := rlg.MethodLimits.RateLimits[i]
		if rl.Current > 0 {
			rl.Current--
		}
	}
}
