package configs

import (
	"strings"
	"time"

	"github.com/DarkIntaqt/cosmic-radiance/internal/utils"
)

// This file parses all environment variables related to configuration.

var (
	// API key to access the Riot Games API
	ApiKeys []string = strings.Split(utils.GetEnvString("API_KEY"), ",")

	// Port cosmic-radiance will be running on
	Port int = utils.GetEnvInt("PORT")

	// Mode cosmic-radiance will be running in, either PATH or PROXY for different request syntax
	Mode RequestMode = validateMode(utils.GetEnvString("MODE"))

	// Timeout for requests
	Timeout time.Duration = parseTimeout(utils.GetSoftEnvString("TIMEOUT", DEFAULT_TIMEOUT_STRING))

	// Size of priority in percent
	PriorityQueueSize float32 = handlePriorityQueueSize()

	// Enable or disable Prometheus metrics
	PrometheusEnabled bool = strings.ToLower(utils.GetSoftEnvString("PROMETHEUS", "OFF")) == "on"
)
