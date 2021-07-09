package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/malyusha/immune-mosru-server/pkg/logger"
	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server"
)

// RecoverOnPanic is the middleware, responsible for handling panics, occurred at runtime.
// It logs panic with given logger and sets http response status to Internal Server Error (500).
func RecoverOnPanic(log logger.Logger) server.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if r := recover(); r != nil {
					debug.PrintStack()
					log.With(logger.Fields{"panic": r}).Error("panic inside of HTTP")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"message":"Internal Server Error"}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}
