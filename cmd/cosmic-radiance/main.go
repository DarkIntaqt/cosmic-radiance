package main

import (
	"github.com/DarkIntaqt/cosmic-radiance/configs"
	"github.com/DarkIntaqt/cosmic-radiance/internal/ratelimiter"
)

// Automatically starts cosmic-radiance
func main() {

	limiter := ratelimiter.NewRateLimiter(configs.Port)
	limiter.Start()

}
