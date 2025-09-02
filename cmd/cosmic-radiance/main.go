package main

import (
	"log"

	"github.com/DarkIntaqt/cosmic-radiance/configs"
	"github.com/DarkIntaqt/cosmic-radiance/internal/ratelimiter"
)

// Automatically starts cosmic-radiance
func main() {

	log.Printf("Cosmic-Radiance v%s on :%d\n", configs.VERSION, configs.Port)

	limiter := ratelimiter.NewRateLimiter(configs.Port)
	limiter.Start()

}
