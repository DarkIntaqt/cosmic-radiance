package configs

import (
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/DarkIntaqt/cosmic-radiance/internal/utils"
)

type RequestMode = bool

const (
	PathMode  RequestMode = false
	ProxyMode RequestMode = true
)

func validateMode(mode string) RequestMode {
	switch strings.ToLower(mode) {
	case "path":
		return PathMode
	case "proxy":
		return ProxyMode
	default:
		panic("Invalid mode, must be 'PATH' or 'PROXY'")
	}
}

func parseTimeout(timeout string) time.Duration {
	duration, err := time.ParseDuration(timeout + "s")
	if err != nil {
		log.Printf("Failed to parse env timeout, falling back to default: %v", err)
		return DEFAULT_TIMEOUT_DURATION
	}

	return duration
}

func handlePriorityQueueSize() float32 {
	limit := utils.GetSoftEnvString("PRIORITY_QUEUE_SIZE", "50")
	value, err := strconv.ParseFloat(limit, 32)
	if err != nil {
		log.Printf("Error parsing PRIORITY_QUEUE_SIZE: %v\n", err)
		return 0.5
	}

	// Clamp the value
	value = math.Min(math.Max(float64(value), 0), 100)

	// Make it a percentage
	return float32(value) / 100
}
