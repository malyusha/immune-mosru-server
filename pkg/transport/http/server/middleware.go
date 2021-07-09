package server

import (
	"net/http"
)

// Middleware represents base application http middleware.
type Middleware func(next http.Handler) http.Handler

// Chain chains middleware list.
func Chain(h http.Handler, m ...Middleware) http.Handler {
	// reversing middleware list:
	// first(second(third(handler)))
	for i := len(m) - 1; i >= 0; i-- {
		h = m[i](h)
	}

	return h
}