package prometheus

import "github.com/prometheus/client_golang/prometheus"

type GenericMetrics struct {
	Info prometheus.Gauge
}

func NewGenericMetrics(constLabels prometheus.Labels) *GenericMetrics {
	m := &GenericMetrics{
		Info: prometheus.NewGauge(
			prometheus.GaugeOpts{
				ConstLabels: constLabels,
				Name:        "info",
				Help:        "Always set to 1 for a running application and is used to track versions.",
			},
		),
	}
	m.Info.Set(1.0)

	prometheus.MustRegister(m.Info)
	return m
}
