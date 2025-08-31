package resource

import (
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/DarkIntaqt/cosmic-radiance/configs"
)

type RateLimit struct {
	Window     time.Duration // seconds
	Limit      int           // allowed requests during the window
	Current    int           // current requests during the window
	LastRefill time.Time
}

func (rlc *RateLimitCategory) Update(limit string, count string, retryAfter *time.Time, applyRetryAfter bool) float64 {
	if limit == "" || count == "" {
		return 0
	}

	limits := strings.Split(limit, ",")
	counts := strings.Split(count, ",")

	peakCapacity := float64(0)
	now := time.Now()

	for i, value := range limits {
		split := strings.SplitN(value, ":", 2)
		if len(split) != 2 {
			continue
		}

		capacity, err := strconv.Atoi(split[0])
		if err != nil {
			continue
		}

		window, err := strconv.Atoi(split[1])
		if err != nil {
			continue
		}

		// check if counts[i] exists, then split and parse the val
		current := 0
		if i < len(counts) {
			countSplit := strings.SplitN(counts[i], ":", 2)
			if len(countSplit) == 2 {
				current, err = strconv.Atoi(countSplit[0])
				if err != nil {
					current = 0
				}
			}
		}

		// update limit if exists, otherwise create
		if i < len(rlc.RateLimits) {

			lastRefill := rlc.RateLimits[i].LastRefill
			currentReqs := rlc.RateLimits[i].Current
			duration := time.Duration(window)*time.Second + 125*time.Millisecond

			// We puddle along the original limit
			// If the limit Riot gave us is higher than ours (which shouldnt be happening), update ours accordingly
			if rlc.RateLimits[i].Limit != capacity && currentReqs != current {
				log.Printf("Updating current rate limit from %d to %d with window %d\n", currentReqs, current, duration)
				currentReqs = current
			}

			if applyRetryAfter && retryAfter != nil && current >= capacity {
				lastRefill = (*retryAfter).Add(-duration)
				currentReqs = 0
			}

			rlc.RateLimits[i] = &RateLimit{
				Window:     duration,
				Limit:      capacity,
				Current:    currentReqs,
				LastRefill: lastRefill,
			}
		} else {

			duration := time.Duration(window) * time.Second
			lastRefill := now
			currentReqs := current

			if applyRetryAfter && retryAfter != nil && current >= capacity {
				lastRefill = (*retryAfter).Add(-duration)
				currentReqs = 0
			}

			rlc.RateLimits = append(rlc.RateLimits, &RateLimit{
				Window:     duration,
				Limit:      capacity,
				Current:    currentReqs,
				LastRefill: lastRefill,
			})
		}

		ratelimit := rlc.RateLimits[i]

		currentPeakCapacity := float64(ratelimit.Limit) / float64(ratelimit.Window) * float64(configs.Timeout)

		if peakCapacity == 0 {
			peakCapacity = currentPeakCapacity
		} else {
			peakCapacity = math.Min(peakCapacity, currentPeakCapacity)
		}
	}

	if applyRetryAfter && retryAfter != nil {
		log.Println("Applying Retry-After:", (*retryAfter).Format(time.RFC3339))
		rlc.LockedUntil = (*retryAfter)
	}

	limitLen := len(limits)
	if limitLen < len(rlc.RateLimits) {
		log.Println("limit", limit)
		log.Println("count", count)
		log.Printf("%+v", rlc.RateLimits)
		log.Printf("Dropping rate limit to %d. Got %d limits, but have %d stored\n", limitLen, limitLen, len(rlc.RateLimits))
		rlc.RateLimits = rlc.RateLimits[:limitLen]
		log.Printf("%+v", rlc.RateLimits)
	}

	return peakCapacity

}
