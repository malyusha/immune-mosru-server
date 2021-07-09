package router

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/malyusha/immune-mosru-server/pkg/errdefs"
	"github.com/malyusha/immune-mosru-server/pkg/logger"
	"github.com/malyusha/immune-mosru-server/pkg/metrics/prometheus"
	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server"
	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server/httputils"
)

type handler struct {
	http.Handler
	logger           logger.Logger
	globalMiddleware []server.Middleware
	metrics          *prometheus.HTTPHandlerMetrics
	routers          []Router
}

func NewRouterHandler(opts ...Option) (*handler, error) {
	h := &handler{
		logger: logger.With(logger.Fields{"package": "router"}),
	}
	evaluateOptions(h, opts...)

	if err := validate(h); err != nil {
		return nil, fmt.Errorf("invalid router handler configuration: %w", err)
	}

	h.Handler = h.createHandler()

	return h, nil
}

func (hl *handler) wrapHandler(h http.Handler) http.Handler {
	// Instrument all routes with metrics, including "Not Found & Not Allowed"
	if hl.metrics != nil {
		h = hl.metrics.InstrumentHandler(h, "all")
	}
	// Add all global middlewares
	h = server.Chain(h, hl.globalMiddleware...)

	return h
}

func (hl *handler) createHandler() http.Handler {
	httpRouter := httprouter.New()
	routers := hl.routers
	log := hl.logger

	log.Debug("registering routers")
	for _, apiRouter := range routers {
		for _, r := range apiRouter.Routes() {
			h := http.Handler(r.Handler())
			// check if router must be wrapped with individual middleware list
			if mwRouter, ok := apiRouter.(ProvidesMiddleware); ok {
				h = server.Chain(h, mwRouter.Middlewares()...)
			}
			if hl.metrics != nil {
				// Wrap each route with metrics
				h = hl.metrics.InstrumentHandler(h, r.Path())
			}

			log.Debugf("registering route %s, %s", r.Method(), r.Path())
			httpRouter.Handler(r.Method(), r.Path(), h)
		}
	}

	// create error handlers
	notFoundError := errdefs.NotFound(errors.New("not found"))
	notFoundHandler := httputils.MakeErrorHandler(notFoundError)

	httpRouter.NotFound = notFoundHandler
	httpRouter.MethodNotAllowed = notFoundHandler

	return hl.wrapHandler(httpRouter)
}

func validate(h *handler) error {
	if len(h.routers) == 0 {
		return errors.New("no routers provided")
	}

	if h.metrics == nil {
		logger.Warn("router metrics not provided. ensure this is expected behaviour")
	}

	return nil
}
