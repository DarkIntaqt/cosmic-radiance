package utils

import (
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DarkIntaqt/cosmic-radiance/ratelimiter/options"
	_ "github.com/joho/godotenv/autoload"
)

// Retrieves the value of the environment variable named by the key.
// Used for required environment variables.
func GetEnvString(key string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		panic("Environment variable " + key + " not set")
	}

	if value == "" {
		panic("Environment variable " + key + " is empty")
	}

	return value
}

// Retrieves the value of the environment variable named by the key, otherwise returns a fallback.
func GetSoftEnvString(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}

	if value == "" {
		return defaultValue
	}

	return value
}

// Retrieves the value of the environment variable named by the key as an integer.
// Used for required environment variables.
func GetEnvInt(key string) int {
	value := GetEnvString(key)

	intValue, err := strconv.Atoi(value)
	if err != nil {
		panic("Environment variable " + key + " is not a valid integer")
	}

	return intValue
}

func ValidateRequestMode() options.CosmicRadianceRequestMode {
	mode := GetEnvString("MODE")

	switch strings.ToLower(mode) {
	case "path":
		return options.PathMode
	case "proxy":
		return options.ProxyMode
	default:
		panic("Invalid mode, must be 'PATH' or 'PROXY'")
	}
}

func HandlePriorityQueueSize() float32 {
	limit := GetSoftEnvString("PRIORITY_QUEUE_SIZE", "50")
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

func HandleDuration(unit string, envName string, defaultDuration time.Duration) time.Duration {
	limit := GetSoftEnvString(envName, "false")
	if limit == "false" {
		return defaultDuration
	}

	duration, err := time.ParseDuration(limit + unit)
	if err != nil {
		log.Printf("Error parsing %s: %v\n", envName, err)
		return defaultDuration
	}

	return duration
}
