package middleware

import (
	"net/http"
	"strings"

	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server"
)

func SetContentType(typ string) server.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", typ)
			next.ServeHTTP(w, r)
		})
	}
}

// ResponseWithKeepAlive responds with Connection: Keep-Alive header for incoming request
// if client requests with http/1.1 or explicitly sets Keep-Alive connection header inside request
// headers.
func ResponseWithKeepAlive(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hConn := strings.ToLower(r.Header.Get("Connection"))
		if hConn == "keep-alive" || hConn == "" {
			w.Header().Set("Connection", "keep-alive")
		}

		next.ServeHTTP(w, r)
	})
}
