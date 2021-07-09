package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/malyusha/immune-mosru-server/pkg/contextutils"
	"github.com/malyusha/immune-mosru-server/pkg/logger"
	"github.com/malyusha/immune-mosru-server/pkg/transport/http/server"
)

const (
	traceIDHeader = "X-Trace-Id"
)

type responseLoggingWriter struct {
	w      http.ResponseWriter
	size   int
	status int
}

func (r *responseLoggingWriter) Header() http.Header {
	return r.w.Header()
}

func (r *responseLoggingWriter) Write(b []byte) (int, error) {
	r.size = r.size + len(b)
	return r.w.Write(b)
}

func (r *responseLoggingWriter) WriteHeader(s int) {
	r.w.WriteHeader(s)
	r.status = s
}

type options struct {
	ignoreAllHeaders bool
	ignoreHeaders    map[string]struct{}
	noLogStatusCodes map[int]struct{}
}

var defaultOptions = &options{
	ignoreHeaders:    make(map[string]struct{}),
	noLogStatusCodes: map[int]struct{}{http.StatusNotFound: {}},
}

type Option func(opts *options)

func IgnoreHeaders(headers ...string) Option {
	return func(opts *options) {
		for _, h := range headers {
			if h == "*" {
				opts.ignoreAllHeaders = true
				return
			}

			opts.ignoreHeaders[strings.ToLower(h)] = struct{}{}
		}
	}
}

func SkipStatusCodes(codes []int) Option {
	return func(opts *options) {
		for _, c := range codes {
			opts.noLogStatusCodes[c] = struct{}{}
		}
	}
}

func evaluateLogOptions(opts ...Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}

	return optCopy
}

// LogRequests logs incoming requests.
func LogRequests(opts ...Option) server.Middleware {
	options := evaluateLogOptions(opts...)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log := logger.WithContext(r.Context())
			start := time.Now()
			writer := &responseLoggingWriter{w: w, status: http.StatusOK}
			next.ServeHTTP(writer, r)
			if _, ok := options.noLogStatusCodes[writer.status]; ok {
				return
			}

			ok := writer.status >= http.StatusOK && writer.status < 400
			fields := logger.Fields{
				"IP":            r.RemoteAddr,
				"URL":           r.URL.String(),
				"method":        r.Method,
				"status_code":   writer.status,
				"response_size": fmt.Sprintf("%.2f KB", float64(writer.size)/1024),
				"started_at":    start,
				"took":          fmt.Sprintf("%.4f sec", time.Since(start).Seconds()),
			}

			var headers map[string]string
			if !options.ignoreAllHeaders {
				headers = make(map[string]string, len(r.Header))
				for k, h := range r.Header {
					// Ignore headers from zerolog (e.g. auth tokens or any other sensitive information)
					if _, ok := options.ignoreHeaders[strings.ToLower(k)]; ok {
						continue
					}

					headers[k] = strings.Join(h, ", ")
				}

				fields["request_headers"] = headers
			}

			if !ok {
				log.With(fields).Error("Response error")
			} else {
				log.With(fields).Info("Response info")
			}
		})
	}
}

// LogFieldsExtractor must extract fields from given context and return
// logger.Fields.
type LogFieldsExtractor func(ctx context.Context) logger.Fields

// PassLoggingFieldsToRequestContext passes zerolog fields to request context
// for later use with logger.WithContext(ctx) to log those fields.
func PassLoggingFieldsToRequestContext(extract LogFieldsExtractor) server.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqCtx := r.Context() // get request context
			// and wrap ctx into zerolog ctx, adding default fields
			r = r.WithContext(logger.NewContext(reqCtx, extract(reqCtx)))
			next.ServeHTTP(w, r)
		})
	}
}

// StringGenerator is used for request-id generation.
type StringGenerator func() string

// GenerateID is a StringGenerator to generate unique ID for each incoming HTTP request.
func GenerateID() string { return contextutils.NewTraceID() }

// Provides trace id to request context.
// First, it looks for trace ID from request headers and if there are no ID set
// it generates random UUID as fallback.
func PassTraceIDToContext(generate StringGenerator) server.Middleware {
	if generate == nil {
		// Fallback to default UUID generate.
		generate = GenerateID
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get trace ID from headers or generate new
			traceID := r.Header.Get(traceIDHeader)
			if traceID == "" {
				traceID = generate()
			}

			r = r.WithContext(contextutils.SetTraceID(r.Context(), traceID))

			next.ServeHTTP(w, r)
		})
	}
}
