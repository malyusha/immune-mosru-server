package router

import (
	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server"
	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server/httputils"
)

// Router defines an interface to specify a group of routes to add to the server.
type Router interface {
	// Routes returns the list of routes to add to the server.
	Routes() []Route
}

// Route defines an individual API route in the server.
type Route interface {
	// handler returns the raw function to create the http handler.
	Handler() httputils.APIFunc
	// Method returns the http method that the route responds to.
	Method() string
	// Path returns the subpath where the route responds to.
	Path() string
}

// ProvidesMiddleware defines interface to specify middleware list to add to router.
type ProvidesMiddleware interface {
	Middlewares() []server.Middleware
}
