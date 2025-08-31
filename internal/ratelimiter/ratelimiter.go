package ratelimiter

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/DarkIntaqt/cosmic-radiance/configs"

	"github.com/DarkIntaqt/cosmic-radiance/internal/metrics"
	"github.com/DarkIntaqt/cosmic-radiance/internal/queue"
)

type RateLimiter struct {
	queueManager    *queue.QueueManager
	incomingChannel chan IncomingRequest
	updateChannel   chan Update
	started         bool
	client          *http.Client
	port            int
	requests        int64
	close           chan struct{}
}

func NewRateLimiter(port int) *RateLimiter {
	queueManager := queue.NewQueueManager()
	incomingChannel := make(chan IncomingRequest)
	updateChannel := make(chan Update)

	return &RateLimiter{
		queueManager:    queueManager,
		incomingChannel: incomingChannel,
		updateChannel:   updateChannel,
		started:         false,
		close:           make(chan struct{}),
		port:            port,
		requests:        1,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

/*
Start the rate limiter. The rate limiter will process incoming requests
in a separate goroutine and ensure that they are handled according to
the defined rate limits.
*/
func (rl *RateLimiter) Start() {
	// Instances shall only run once to ensure thread safety
	if rl.started {
		return
	}
	rl.started = true

	// Add a cancel function
	ctx, cancel := context.WithCancel(context.Background())

	if configs.PrometheusEnabled {
		metrics.InitMetrics()
	}

	// Start the main loop in a goroutine
	go func() {

		log.Println("Starting main loop")

		rl.mainLoop(ctx)

	}()

	// Create the http proxy
	proxy := &http.Server{
		Addr:    fmt.Sprintf(":%d", rl.port),
		Handler: rl,
	}

	// Serve the http proxy
	go func() {

		log.Println("Starting proxy")

		if err := proxy.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Proxy crashed: %v\n", err)
		}

	}()

	// Listen to TERM signals, if so, shut down
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	// Printing an empty line to distinct between before and after the shutdown. Also prevents ^C to be visible in another log message
	println("")
	log.Println("Shutting down...")

	// Cancel the context to stop the goroutine
	cancel()

	// Create a deadline after which the program would force exit if not shutdown successfully. 30 seconds are very generous
	stop, cancelDeadline := context.WithDeadline(context.Background(), time.Now().Add(30*time.Second))
	defer cancelDeadline()

	// Shutdown the http proxy
	if err := proxy.Shutdown(stop); err != nil {
		log.Printf("Proxy forced to shutdown: %v\n", err)
	} else {
		log.Println("Bye bye from the proxy")
	}

	close(rl.incomingChannel)
	close(rl.updateChannel)

	// Waiting for the goroutine to finish or the context to be done
	// TODO: I don't know if there *could* be a race condition here causing the proxy to stop with ctx.stop and the goroutine not finishing.
	select {
	case <-rl.close:
	case <-stop.Done():
		log.Println("The goroutine didn't stop in time, we forcefully shutting it down now")
	}

	log.Println("Bye bye from the main thread")
}

/*
INTERNAL:
Append incoming requests and process the queues
*/
func (rl *RateLimiter) mainLoop(ctx context.Context) {
	// Process incoming requests

	ticker := time.NewTicker(30 * time.Second)
	metricsTicker := time.NewTicker(1 * time.Second)
	if !configs.PrometheusEnabled {
		metricsTicker.Stop()
	}

	for {
		select {
		case req := <-rl.incomingChannel:
			rl.requests++ // increase requests, just to be sure
			rl.handleIncomingRequest(req)

		case update := <-rl.updateChannel:
			rl.handleUpdateRequest(update)
			if update.RetryAfter == nil {
				rl.queueManager.AdjustQueueSize()
			}

		case <-ctx.Done():
			ticker.Stop()
			if configs.PrometheusEnabled {
				metricsTicker.Stop()
			}
			// Drain all queues before shutting down
			rl.queueManager.Drain()

			// This is intended and absolutely necessary
			log.Println("Bye bye from the main loop")

			// Send data to notify the main thread that worker has finished
			rl.close <- struct{}{}
			return
		case <-metricsTicker.C:
			metrics.UpdateQueueSizes(rl.queueManager)

		case <-ticker.C:
			size := int64(0)
			for _, queue := range rl.queueManager.PriorityQueues {
				size += queue.Count()
			}
			for _, queue := range rl.queueManager.Queues {
				size += queue.Count()
			}

			rl.requests = size
			rl.queueManager.CleanUp()

		default:
			if rl.requests == 0 {
				// Slow CPU cycles
				// Still check everything in case something went wrong
				time.Sleep(250 * time.Millisecond)
			}
			rl.refillRateLimits()
			rl.processQueues(rl.queueManager.PriorityQueues)
			rl.processQueues(rl.queueManager.Queues)

		}

	}
}
