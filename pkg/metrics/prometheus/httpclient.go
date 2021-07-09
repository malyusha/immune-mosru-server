package prometheus

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HTTPClientMetrics struct {
	Requests *prometheus.CounterVec
	Duration *prometheus.HistogramVec
	InFlight prometheus.Gauge
}

func NewHTTPClientMetrics(constLabels prometheus.Labels) *HTTPClientMetrics {
	m := &HTTPClientMetrics{
		Requests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				ConstLabels: constLabels,
				Name:        "http_client_requests_total",
				Help:        "Number of HTTP client requests.",
			},
			[]string{"handler", "code", "method"},
		),
		Duration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				ConstLabels: constLabels,
				Name:        "http_client_requests_duration_seconds",
				Help:        "Duration of HTTP client requests.",
				Buckets:     DefaultDurationBuckets,
			},
			[]string{"handler", "code", "method"},
		),
		InFlight: prometheus.NewGauge(
			prometheus.GaugeOpts{
				ConstLabels: constLabels,
				Name:        "http_client_requests_in_flight_total",
				Help:        "Number of client HTTP requests in flight.",
			},
		),
	}

	prometheus.MustRegister(
		m.Requests,
		m.Duration,
		m.InFlight,
	)

	return m
}

// InstrumentRoundTripper instruments http.Roundtripper with metrics m.
func (m *HTTPClientMetrics) InstrumentRoundTripper(rt http.RoundTripper, handler string) http.RoundTripper {
	rt = promhttp.InstrumentRoundTripperCounter(m.Requests.MustCurryWith(prometheus.Labels{"handler": handler}), rt)
	rt = promhttp.InstrumentRoundTripperDuration(m.Duration.MustCurryWith(prometheus.Labels{"handler": handler}), rt)
	rt = promhttp.InstrumentRoundTripperInFlight(m.InFlight, rt)
	return rt
}
