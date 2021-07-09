package httputils

import (
	"context"
	"fmt"
	"mime"
	"net/http"

	"github.com/malyusha/immune-mosru-server/pkg/errdefs"
	"github.com/malyusha/immune-mosru-server/pkg/logger"
)

// RouteParams represents map of route's path parameters.
type RouteParams map[string]string

// APIFunc is an adapter to allow the use of ordinary functions as Docker API endpoints.
// Any function that has the appropriate signature can be registered as an API endpoint (e.g. getVersion).
type APIFunc func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// ServeHTTP implements http.Handler interface and handles APIFunc error if any occurred.
func (f APIFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	WrapAPIFunc(f).ServeHTTP(w, r)
}

// WrapAPIFunc creates http.Handler from given APIFunc.
func WrapAPIFunc(f APIFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := f(r.Context(), w, r)
		if err == nil {
			return
		}

		// if APIFunc returns error we must process it.
		statusCode := errdefs.GetHTTPErrorStatusCode(err)
		if statusCode >= 500 {
			logger.WithContext(r.Context()).Errorf("Handler for %s %s returned error: %v", r.Method, r.URL.Path, err)
		}

		// dynamically create http handler for error
		handleError := MakeErrorHandler(err)
		// Handle error if occurred
		handleError(w, r)
	})
}

// CheckForJSON makes sure that the request's Content-Type is application/json.
func CheckForJSON(r *http.Request) error {
	ct := r.Header.Get("Content-Type")

	// No Content-Type header is ok as long as there's no Body
	if ct == "" {
		if r.Body == nil || r.ContentLength == 0 {
			return nil
		}
	}

	// Otherwise it better be json
	if matchesContentType(ct, "application/json") {
		return nil
	}
	return errdefs.InvalidParameter(fmt.Errorf("Content-Type specified (%s) must be 'application/json'", ct))
}

type baseErr struct{ Message string `json:"message"` }
type validationError struct {
	Message string                `json:"message"`
	Fields  errdefs.InvalidFields `json:"fields"`
}

// MakeErrorHandler makes an HTTP handler that decodes an error and
// returns it in the response.
func MakeErrorHandler(err error) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		statusCode := errdefs.GetHTTPErrorStatusCode(err)

		var respError interface{}
		switch {
		case errdefs.IsUnavailable(err):
			respError = baseErr{Message: "Service Unavailable"}
		case errdefs.IsNotImplemented(err):
			respError = baseErr{Message: "Not Implemented"}
		case errdefs.IsSystem(err), errdefs.IsUnknown(err):
			respError = baseErr{Message: "Internal Server Error"}
		case errdefs.IsValidation(err):
			respError = validationError{Message: err.Error(), Fields: err.(errdefs.ErrValidation).Fields()}
		default:
			respError = baseErr{Message: err.Error()}
		}

		_ = WriteJSON(w, statusCode, respError)
	}
}

// HasBoolVar checks whether http request query parameters contain given key.
func HasBoolVar(r *http.Request, key string) bool {
	values := r.URL.Query()
	_, ok := values[key]

	return ok
}

// matchesContentType validates the content type against the expected one
func matchesContentType(contentType, expectedType string) bool {
	mimetype, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		logger.Errorf("Error parsing media type: %s error: %v", contentType, err)
	}
	return err == nil && mimetype == expectedType
}
