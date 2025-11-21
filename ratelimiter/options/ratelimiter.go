package options

import (
	"time"
)

type CosmicRadianceRequestMode = bool

// Request modes for the Proxy. The path mode ignores the host, the proxy mode uses the host to get the Riot Games API URL.
const (
	PathMode  CosmicRadianceRequestMode = false
	ProxyMode CosmicRadianceRequestMode = true
)

type RateLimiterOptions struct {
	ApiKeys              []string
	Port                 int
	RequestMode          CosmicRadianceRequestMode
	Timeout              time.Duration
	PriorityQueueSize    float32
	PrometheusEnabled    bool
	PollingInterval      time.Duration
	AdditionalWindowSize time.Duration
}

func ValidateRateLimiterOptions(opts *RateLimiterOptions) {

	if len(opts.ApiKeys) == 0 {
		panic("Prove an API key")
	}

	if opts.Port <= 0 || opts.Port > 65535 {
		panic("Invalid port number")
	}

	if opts.Timeout <= 0 {
		panic("Timeout must be greater than 0")
	}

	if opts.PriorityQueueSize < 0 || opts.PriorityQueueSize > 1 {
		panic("Priority queue size must be between 0 and 1")
	}

	if opts.PollingInterval <= 0 {
		panic("Polling interval must be greater than 0")
	}

	if opts.AdditionalWindowSize < 0 {
		panic("Additional window size must be greater than or equal to 0")
	}

}
