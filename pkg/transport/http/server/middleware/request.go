package middleware

import (
	"net/http"

	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server"
)

func LimitRequestBody(maxBytes int64) server.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}