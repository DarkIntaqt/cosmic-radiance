package configs

import (
	"fmt"
	"strings"
	"time"

	"github.com/DarkIntaqt/cosmic-radiance/internal/utils"
)

// This file parses all environment variables related to configuration.

// API key to access the Riot Games API
var ApiKeys []string = strings.Split(utils.GetEnvString("API_KEY"), ",")

// Port cosmic-radiance will be running on
var Port int = utils.GetEnvInt("PORT")

// Mode cosmic-radiance will be running in, either PATH or PROXY for different request syntax
var RequestMode requestMode = validateMode(utils.GetEnvString("MODE"))

// Timeout for requests
var Timeout time.Duration = parseTimeout(utils.GetSoftEnvString("TIMEOUT", fmt.Sprintf("%d", int(DEFAULT_INCOMING_REQUEST_TIMEOUT.Seconds()))))

// Size of priority in percent
var PriorityQueueSize float32 = handlePriorityQueueSize()

// Enable or disable Prometheus metrics
var PrometheusEnabled bool = strings.ToLower(utils.GetSoftEnvString("PROMETHEUS", "OFF")) == "on"
