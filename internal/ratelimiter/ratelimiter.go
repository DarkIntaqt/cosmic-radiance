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
	queueManager *queue.QueueManager

	incomingChannel chan IncomingRequest
	updateChannel   chan Update
	refundChannel   chan Refund
	stopSignal      chan os.Signal
	close           chan struct{}

	started bool

	client   *http.Client
	port     int
	requests int64
}

func NewRateLimiter(port int) *RateLimiter {

	if configs.MAX_UTILIZATION_FACTOR <= 0 || configs.MAX_UTILIZATION_FACTOR > 1 {
		panic("Invalid MAX_UTILIZATION_FACTOR")
	}

	queueManager := queue.NewQueueManager()

	stopSignal := make(chan os.Signal, 1)
	signal.Notify(stopSignal, syscall.SIGINT, syscall.SIGTERM)

	return &RateLimiter{
		queueManager: queueManager,
		stopSignal:   stopSignal,
		started:      false,
		close:        make(chan struct{}),
		port:         port,
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

	log.Printf("Running Cosmic-Radiance v%s on :%d\n", configs.VERSION, configs.Port)

	// Create all channels,
	rl.incomingChannel = make(chan IncomingRequest)
	rl.updateChannel = make(chan Update)
	rl.refundChannel = make(chan Refund)

	// Add a cancel function
	ctx, cancelCtx := context.WithCancel(context.Background())

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
	<-rl.stopSignal

	// Printing an empty line to distinct between before and after the shutdown. Also prevents ^C to be visible in another log message
	println("")
	log.Println("Shutting down...")

	// Cancel the context to stop the goroutine
	cancelCtx()

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
	close(rl.refundChannel)

	// Waiting for the goroutine to finish or the context to be done
	// TODO: I don't know if there *could* be a race condition here causing the proxy to stop with ctx.stop and the goroutine not finishing.
	select {
	case <-rl.close:
	case <-stop.Done():
		log.Println("The goroutine didn't stop in time, we forcefully shutting it down now")
	}

	log.Println("Bye bye from the main thread")
}

func (rl *RateLimiter) Stop() {
	if rl.started {
		rl.stopSignal <- syscall.SIGTERM
	}
}

/*
INTERNAL:
Append incoming requests and process the queues
*/
func (rl *RateLimiter) mainLoop(ctx context.Context) {
	// Process incoming requests

	// Tickers for irrelevant tasks such as clean up processes (free memory)
	cleanUpTicker := time.NewTicker(30 * time.Second)

	metricsTicker := time.NewTicker(1 * time.Second)
	if !configs.PrometheusEnabled {
		metricsTicker.Stop()
	}

	pollingTicker := time.NewTicker(configs.PollingInterval)

	for {
		select {
		case req := <-rl.incomingChannel:
			rl.handleIncomingRequest(req)

		case update := <-rl.updateChannel:
			rl.handleUpdateRequest(update)
			if update.RetryAfter == nil {
				rl.queueManager.AdjustQueueSize()
			}

		case refund := <-rl.refundChannel:
			rl.handleRefund(refund)

		case <-ctx.Done():
			cleanUpTicker.Stop()
			if configs.PrometheusEnabled {
				metricsTicker.Stop()
			}
			pollingTicker.Stop()

			// Drain all queues before shutting down
			rl.queueManager.Drain()

			// This is intended and absolutely necessary
			log.Println("Bye bye from the main loop")

			// Send data to notify the main thread that worker has finished
			rl.close <- struct{}{}
			return
		case <-metricsTicker.C:
			metrics.UpdateQueueSizes(rl.queueManager)

		case <-cleanUpTicker.C:
			rl.queueManager.CleanUp()

		case <-pollingTicker.C:
			rl.refillRateLimits()
			rl.processQueues(rl.queueManager.PriorityQueues)
			rl.processQueues(rl.queueManager.Queues)
		}

	}
}
