package ratelimiter

import (
	"fmt"
	"sync"

	"github.com/DarkIntaqt/cosmic-radiance/internal/ratelimiter"
)

type cosmicRadiance struct {
	instance *ratelimiter.RateLimiter
	running  bool
	mu       sync.Mutex
}

// Initializes a new instance of cosmic-radiance
func Init(port int) *cosmicRadiance {
	instance := ratelimiter.NewRateLimiter(port)

	return &cosmicRadiance{
		instance: instance,
		running:  false,
		mu:       sync.Mutex{},
	}
}

// Starts a non-blocking cosmic-radiance instance if it is not already running
func (cr *cosmicRadiance) Start() error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if cr.running {
		return fmt.Errorf("cosmic-radiance is already running")
	}

	cr.running = true
	go cr.instance.Start()

	return nil
}

// Starts a blocking cosmic-radiance instance if it is not already running
func (cr *cosmicRadiance) Run() error {
	cr.mu.Lock()

	if cr.running {
		cr.mu.Unlock()
		return fmt.Errorf("cosmic-radiance is already running")
	}

	cr.running = true
	cr.mu.Unlock()
	cr.instance.Start()

	return nil
}

// Stops a running cosmic-radiance instance. This function is non-blocking, cosmic-radiance might still be shutting down when this function returns
func (cr *cosmicRadiance) Stop() error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if !cr.running {
		return fmt.Errorf("cosmic-radiance is not running")
	}

	cr.instance.Stop()
	cr.running = false

	return nil
}
