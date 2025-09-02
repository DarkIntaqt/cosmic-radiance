package configs

import "time"

// Incoming request timeout configs. This variable can be changed with a .env file.
const DEFAULT_INCOMING_REQUEST_TIMEOUT = 10 * time.Second

// Interval in which the rate limit is checked against the Riot Games API.
// It is very unlikely that rate limits change, but it should be accounted for.
const RATELIMIT_UPDATE_INTERVAL = time.Minute * 1

// Time to wait between a RATELIMIT_UPDATE_INTERVAL before requesting the next update.
const UPDATE_GRATUITY = time.Millisecond * 200

// Duration after which a inactive queue gets deleted.
const QUEUE_INACTIVITY = time.Minute * 10
