package metrics

import "github.com/prometheus/client_golang/prometheus"

// Namespace used for all metrics of this service
const metricNamespace = "vodpacker"

type Metrics struct {
	HttpRequestsMeter *prometheus.CounterVec
	HttpDurationMeter *prometheus.HistogramVec
	// TODO logic operation metrics (using tags: operationName, result, resultValue)
}

// Counter for total HTTP requests (automatically labelled by chi middleware)
var (
	httpRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: metricNamespace,
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests processed, labeled by path and status code.",
		},
		[]string{"path", "method", "code"},
	)

	// Histogram for request latency
	httpDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: metricNamespace,
			Name:      "http_request_duration_seconds",
			Help:      "Latency of HTTP requests.",
			Buckets:   prometheus.DefBuckets,
		},
		[]string{"path", "method"},
	)

	// TODO do this better, need to use tags not this
	// Custom counter: how many successful pack operations we performed
	//packSuccesses = prometheus.NewCounter(
	//	prometheus.CounterOpts{
	//		Namespace: metricNamespace,
	//		Name:      "pack_success_total",
	//		Help:      "Number of successful video‑segment packing operations.",
	//	},
	//)
)

func InitializeMetrics() Metrics {
	metrics := Metrics{
		HttpRequestsMeter: httpRequests,
		HttpDurationMeter: httpDuration,
	}

	register(metrics)

	return metrics
}

func register(metrics Metrics) {
	// Here be dragons!
	// TODO encapsulate side effects if possible, but maybe it isn't very Go lang
	prometheus.MustRegister(metrics.HttpRequestsMeter, metrics.HttpDurationMeter)
}
