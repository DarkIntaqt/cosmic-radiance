package ratelimiter

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/DarkIntaqt/cosmic-radiance/internal/metrics"
	"github.com/DarkIntaqt/cosmic-radiance/internal/request"
	"github.com/DarkIntaqt/cosmic-radiance/internal/schema"
	"github.com/DarkIntaqt/cosmic-radiance/ratelimiter/options"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Serve http requests
func (rl *RateLimiter) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path

	prometheusEnabled := rl.opts.PrometheusEnabled
	// Serve prometheus metrics
	if prometheusEnabled && path == "/metrics" {
		promhttp.Handler().ServeHTTP(w, r)
		return
	}

	var syntax *schema.Syntax

	// Determine the endpoints by using the proxy mode
	if rl.opts.RequestMode == options.ProxyMode {
		schema, err := schema.NewProxySyntax(r.URL.Host, path)
		if err != nil {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		syntax = schema
	} else {
		schema, err := schema.NewPathSyntax(path)
		if err != nil {
			w.Header().Set("Retry-After", "60")
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}
		syntax = schema
	}

	priority := request.NormalPriority
	if r.Header.Get("X-Priority") == "high" {
		priority = request.HighPriority
	}

	timeout := rl.opts.Timeout
	// Create a new request
	req := request.NewRequest(timeout)

	// Don't leave dangling channels open
	// defer close(req.Response)

	// Enqueue request
	rl.incomingChannel <- IncomingRequest{
		Request:  req,
		Syntax:   syntax,
		Priority: priority,
	}

	// add one second on top to not drop requests which should've been successful
	ctx, cancel := context.WithTimeout(context.Background(), timeout+5*time.Second)
	defer cancel()

	select {
	// Handle client cancellations
	// TODO: remove the request from queue
	case <-r.Context().Done():
		if prometheusEnabled {
			metrics.UpdateResponseCodes(-1, syntax.Platform, syntax.Endpoint, 499)
		}

	// The request timed out (internally)
	case <-ctx.Done():
		// fmt.Println("ctx cancelled")
		http.Error(w, "Request dropped due to timeout", http.StatusTooManyRequests)
		if prometheusEnabled {
			metrics.UpdateResponseCodes(-1, syntax.Platform, syntax.Endpoint, 408)
		}

	// The request is allowed to be executed
	case response := <-req.Response:
		// Record the start time, add a small buffer to avoid hitting the next window on accident
		// startTime := time.Now().Add(time.Millisecond * -5)

		if response.KeyId == request.RequestFailed {
			if response.RetryAfter != nil {
				w.Header().Set("Retry-After", fmt.Sprintf("%d", int(time.Until(*response.RetryAfter).Round(time.Second).Seconds())))
			}
			// fmt.Println("timeout exceeded")
			http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
			if prometheusEnabled {
				metrics.UpdateResponseCodes(response.KeyId, syntax.Platform, syntax.Endpoint, 430)
			}
			return
		}

		riotApiRequest, err := rl.riotApiRequest(syntax.Platform, syntax.Method, r.URL.Query(), response.KeyId)
		if err != nil {
			log.Println(err)
			w.Header().Set("Retry-After", "0")
			http.Error(w, "Failed to make API request", http.StatusInternalServerError)

			if prometheusEnabled {
				metrics.UpdateResponseCodes(response.KeyId, syntax.Platform, syntax.Endpoint, 500)
			}
			// rl.refundRequest(syntax, priority, response.KeyId, startTime)
			return
		}

		// Report prometheus statistics, if enabled
		if prometheusEnabled {
			metrics.UpdateResponseCodes(response.KeyId, syntax.Platform, syntax.Endpoint, riotApiRequest.StatusCode)
		}
		defer riotApiRequest.Body.Close()

		if riotApiRequest.StatusCode == http.StatusTooManyRequests || (response.Update && riotApiRequest.StatusCode == http.StatusOK) {
			rl.updateRatelimits(syntax, riotApiRequest, response.KeyId, priority)
		}
		// else if riotApiRequest.StatusCode >= 500 {
		// 	rl.refundRequest(syntax, priority, response.KeyId, startTime)
		// }

		// Copy relevant headers from Riot API response to our response
		importantHeaders := []string{
			"Content-Type", "Content-Encoding", "Content-Length",
			"X-App-Rate-Limit-Count", "X-App-Rate-Limit", "X-Method-Rate-Limit-Count", "X-Method-Rate-Limit", "Retry-After", "X-Rate-Limit-Type",
		}
		for _, key := range importantHeaders {
			if values := riotApiRequest.Header[key]; len(values) > 0 {
				w.Header()[key] = values
			}
		}

		w.Header().Set("X-Key", fmt.Sprintf("%d", response.KeyId+1))

		// Write response 1:1 to keep gzip
		w.WriteHeader(riotApiRequest.StatusCode)
		if _, err := io.Copy(w, riotApiRequest.Body); err != nil {
			log.Printf("Error writing response: %v", err)
		}
	}
}

// func (rl *RateLimiter) refundRequest(syntax *schema.Syntax, priority request.Priority, keyId int, timestamp time.Time) {
// 	rl.refundChannel <- Refund{
// 		Syntax:      syntax,
// 		Priority:    priority,
// 		KeyId:       keyId,
// 		RequestTime: timestamp,
// 	}
// }
