package configs

import "time"

// Incoming request timeout configs
const DEFAULT_TIMEOUT_STRING = "10"
const DEFAULT_TIMEOUT_DURATION = 10 * time.Second

const RATELIMIT_UPDATE_INTERVAL = time.Minute * 1
const UPDATE_GRATUITY = time.Millisecond * 200

const QUEUE_INACTIVITY = time.Minute * 10
