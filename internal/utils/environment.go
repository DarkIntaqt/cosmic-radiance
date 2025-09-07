package utils

import (
	"os"
	"strconv"

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
