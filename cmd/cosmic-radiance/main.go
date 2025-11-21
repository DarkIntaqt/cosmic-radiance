package main

import (
	"strings"

	"github.com/DarkIntaqt/cosmic-radiance/configs"
	"github.com/DarkIntaqt/cosmic-radiance/ratelimiter/options"

	"github.com/DarkIntaqt/cosmic-radiance/internal/ratelimiter"
	"github.com/DarkIntaqt/cosmic-radiance/internal/utils"
)

// Automatically starts cosmic-radiance
func main() {

	limiter := ratelimiter.NewRateLimiter(&options.RateLimiterOptions{
		ApiKeys:           strings.Split(utils.GetEnvString("API_KEY"), ","),
		Port:              utils.GetEnvInt("PORT"),
		RequestMode:       utils.ValidateRequestMode(),
		Timeout:           utils.HandleDuration("s", "TIMEOUT", configs.DEFAULT_INCOMING_REQUEST_TIMEOUT),
		WebserverEnabled:  true,
		PriorityQueueSize: utils.HandlePriorityQueueSize(),
		PrometheusEnabled: strings.ToLower(utils.GetSoftEnvString("PROMETHEUS", "OFF")) == "on",
		PollingInterval:   utils.HandleDuration("ms", "POLLING_INTERVAL", configs.DEFAULT_POLLING_INTERVAL),
		AdditionalWindowSize: utils.HandleDuration("ms", "ADDITIONAL_WINDOW_SIZE",
			configs.DEFAULT_ADDITIONAL_WINDOW_SIZE),
	})
	limiter.Start()

}
