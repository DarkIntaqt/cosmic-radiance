package metrics

import (
	"strconv"

	"github.com/DarkIntaqt/cosmic-radiance/internal/queue"
	"github.com/DarkIntaqt/cosmic-radiance/internal/schema"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	keyResponseCodes *prometheus.CounterVec
	queueSize        *prometheus.GaugeVec
	queueFilled      *prometheus.GaugeVec
	queueCount       *prometheus.GaugeVec
)

func InitMetrics() {
	keyResponseCodes = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "key_response_code_count",
			Help: "Number of responses by key ID, platform, endpoint and response code",
		},
		[]string{"key_id", "platform", "endpoint", "response_code"},
	)
	queueSize = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queue_max_size",
			Help: "Maximum size of a queue",
		},
		[]string{"platform", "endpoint", "priority"},
	)
	queueFilled = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queue_currently_filled",
			Help: "Current amount of requests in the queue",
		},
		[]string{"platform", "endpoint", "priority"},
	)
	queueCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "queue_count",
			Help: "Current count of queues",
		},
		[]string{"priority"},
	)
}

func UpdateResponseCodes(keyId int, platform string, endpoint string, responseCode int) {
	code := strconv.Itoa(responseCode)
	endpoint = "/" + endpoint

	if keyId >= 0 {
		keyResponseCodes.WithLabelValues(strconv.Itoa(keyId+1), platform, endpoint, code).Inc()
	} else {
		keyResponseCodes.WithLabelValues("NO-KEY", platform, endpoint, code).Inc()
	}

}

func UpdateQueueSizes(qm *queue.QueueManager) {

	normal := 0
	priority := 0

	for platform, methods := range schema.AllowedPattern {
		for _, endpoint := range methods {
			id := endpoint.Id
			method := "/" + endpoint.Method

			if queue, Ok := qm.Queues[id]; Ok {
				queueSize.WithLabelValues(platform, method, "normal").Set(float64(queue.Size()))
				queueFilled.WithLabelValues(platform, method, "normal").Set(float64(queue.Count()))
				normal++
			}

			if queue, Ok := qm.PriorityQueues[id]; Ok {
				queueSize.WithLabelValues(platform, method, "high").Set(float64(queue.Size()))
				queueFilled.WithLabelValues(platform, method, "high").Set(float64(queue.Count()))
				priority++
			}
		}
	}

	queueCount.WithLabelValues("normal").Set(float64(normal))
	queueCount.WithLabelValues("high").Set(float64(priority))

}
