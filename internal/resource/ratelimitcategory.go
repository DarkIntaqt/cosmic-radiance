package resource

import "time"

type RateLimitCategory struct {
	LockedUntil          time.Time
	RateLimits           []*RateLimit
	AdditionalWindowSize *time.Duration
	Timeout              *time.Duration
}
