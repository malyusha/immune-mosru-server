package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPHandlerMetrics struct {
	EmptyRequests     *prometheus.CounterVec
	Requests          *prometheus.CounterVec
	Duration          *prometheus.HistogramVec
	InFlight          prometheus.Gauge
	RequestSize       *prometheus.HistogramVec
	ResponseSize      *prometheus.HistogramVec
	TimeToWriteHeader *prometheus.HistogramVec
}

func NewHTTPHandlerMetrics(constLabels prometheus.Labels) *HTTPHandlerMetrics {
	m := &HTTPHandlerMetrics{
		Requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				ConstLabels: constLabels,
				Name:        "http_requests_total",
				Help:        "Total number of HTTP request",
			},
			[]string{"handler", "code", "method"},
		),
		EmptyRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				ConstLabels: constLabels,
				Name:        "http_empty_requests",
				Help:        "Total number of empty HTTP request (empty request logic depends on handler)",
			},
			[]string{"handler"},
		),
		Duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				ConstLabels: constLabels,
				Name:        "http_duration_seconds",
				Help:        "HTTP duration.",
				Buckets:     DefaultDurationBuckets,
			},
			[]string{"handler", "code", "method"},
		),
		InFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				ConstLabels: constLabels,
				Name:        "http_requests_in_flight_total",
				Help:        "Number of HTTP requests in flight.",
			},
		),
		RequestSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				ConstLabels: constLabels,
				Name:        "http_request_size_bytes",
				Help:        "Size of HTTP request.",
				Buckets:     DefaultSizeBuckets,
			},
			[]string{"handler", "code", "method"},
		),
		ResponseSize: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				ConstLabels: constLabels,
				Name:        "http_response_size_bytes",
				Help:        "Size of HTTP response.",
				Buckets:     DefaultSizeBuckets,
			},
			[]string{"handler", "code", "method"},
		),
		TimeToWriteHeader: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				ConstLabels: constLabels,
				Name:        "http_time_to_write_headers_seconds",
				Help:        "Time to write HTTP headers.",
				Buckets:     DefaultDurationBuckets,
			},
			[]string{"handler", "code", "method"},
		),
	}

	prometheus.MustRegister(
		m.Requests,
		m.EmptyRequests,
		m.Duration,
		m.InFlight,
		m.RequestSize,
		m.ResponseSize,
		m.TimeToWriteHeader,
	)

	return m
}

// InstrumentHandler instruments http.Handler with metrics m.
func (m *HTTPHandlerMetrics) InstrumentHandler(h http.Handler, name string) http.Handler {
	h = promhttp.InstrumentHandlerCounter(m.Requests.MustCurryWith(prometheus.Labels{"handler": name}), h)
	h = promhttp.InstrumentHandlerDuration(m.Duration.MustCurryWith(prometheus.Labels{"handler": name}), h)
	h = promhttp.InstrumentHandlerInFlight(m.InFlight, h)
	h = promhttp.InstrumentHandlerRequestSize(m.RequestSize.MustCurryWith(prometheus.Labels{"handler": name}), h)
	h = promhttp.InstrumentHandlerResponseSize(m.ResponseSize.MustCurryWith(prometheus.Labels{"handler": name}), h)
	h = promhttp.InstrumentHandlerTimeToWriteHeader(m.TimeToWriteHeader.MustCurryWith(prometheus.Labels{"handler": name}), h)
	return h
}
