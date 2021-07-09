package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics represents set of pre-defined metrics to use across service.
type Metrics struct {
	Generic *GenericMetrics
	// HTTP handler metrics
	HTTPHandler *HTTPHandlerMetrics
	// HTTP client metrics
	HTTPClient *HTTPClientMetrics
}

var (
	DefaultDurationBuckets = []float64{.001, .0025, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 25, 60}
	DefaultSizeBuckets     = prometheus.ExponentialBuckets(16, 2, 24)
)

// New creates new metrics instance.
// Accepts constant labels, which would be used for all of metrics.
func New(labels prometheus.Labels) *Metrics {
	constLabels := labels

	m := &Metrics{
		Generic:     NewGenericMetrics(constLabels),
		HTTPHandler: NewHTTPHandlerMetrics(constLabels),
		HTTPClient:  NewHTTPClientMetrics(constLabels),
	}

	return m
}