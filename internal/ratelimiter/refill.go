package ratelimiter

import (
	"time"
)

func (rl *RateLimiter) refillRateLimits() {
	keys := rl.queueManager.RateLimitCategories
	for i := range keys {
		categories := keys[i]
		for ckey, category := range categories {
			for j := range category.RateLimits {
				limit := category.RateLimits[j]
				// Check if the limit is expired
				// Refilling does not look super beautiful here because I am not working with pointers but raw values
				if limit.Current > 0 && limit.LastRefill.Add(limit.Window).Before(time.Now()) {
					keys[i][ckey].RateLimits[j].Current = 0
					keys[i][ckey].RateLimits[j].LastRefill = time.Now()
				}
			}
		}
	}
}
