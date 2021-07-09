package router

import (
	"github.com/malyusha/immune-mosru-server/pkg/metrics/prometheus"
	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server"
)

type Option func(h *handler)

func evaluateOptions(h *handler, o ...Option) {
	for _, opt := range o {
		opt(h)
	}
}

func WithGlobalMiddleware(m ...server.Middleware) Option {
	return func(h *handler) {
		h.globalMiddleware = append(h.globalMiddleware, m...)
	}
}

func WithRouter(r ...Router) Option {
	return func(h *handler) {
		h.routers = append(h.routers, r...)
	}
}

func WithMetrics(metrics *prometheus.HTTPHandlerMetrics) Option {
	return func(h *handler) {
		h.metrics = metrics
	}
}
