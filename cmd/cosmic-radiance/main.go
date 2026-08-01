package main

import (
	"fmt"
	"strings"

	"github.com/DarkIntaqt/cosmic-radiance/configs"
	"github.com/DarkIntaqt/cosmic-radiance/ratelimiter/options"

	"github.com/DarkIntaqt/cosmic-radiance/internal/ratelimiter"
	"github.com/DarkIntaqt/cosmic-radiance/internal/utils"
)

// Automatically starts cosmic-radiance
func main() {
	limiter := ratelimiter.NewRateLimiter(&options.RateLimiterOptions{
		ApiKeys:           getApiKeys(),
		Port:              utils.GetEnvInt("PORT"),
		RequestMode:       utils.ValidateRequestMode(),
		Timeout:           utils.HandleDuration("s", "TIMEOUT", configs.DEFAULT_INCOMING_REQUEST_TIMEOUT),
		PriorityQueueSize: utils.HandlePriorityQueueSize(),
		PrometheusEnabled: strings.ToLower(utils.GetSoftEnvString("PROMETHEUS", "OFF")) == "on",
		PollingInterval:   utils.HandleDuration("ms", "POLLING_INTERVAL", configs.DEFAULT_POLLING_INTERVAL),
		AdditionalWindowSize: utils.HandleDuration("ms", "ADDITIONAL_WINDOW_SIZE",
			configs.DEFAULT_ADDITIONAL_WINDOW_SIZE),
		UserAgent: utils.GetSoftEnvString("USER_AGENT", configs.DEFAULT_USER_AGENT),
	})

	limiter.Start()
}

func getApiKeys() []options.KeyKV {
	keys := strings.Split(utils.GetEnvString("API_KEY"), ",")
	i := 0
	apiKeys := make([]options.KeyKV, len(keys))
	for _, key := range keys {
		name := fmt.Sprintf("Key %d", i+1)
		splits := strings.Split(key, "=")
		if len(splits) == 2 && strings.Contains(splits[1], "RGAPI") {
			name = splits[0]
			key = splits[1]
		}

		apiKeys[i] = options.KeyKV{
			ApiKey: key,
			Name:   name,
		}
		i++
	}
	return apiKeys
}
