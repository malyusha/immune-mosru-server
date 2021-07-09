package httputils

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/malyusha/immune-mosru-server/pkg/errdefs"
)

// DecodeJSONRequest tries to decode body of http request into given dst.
// Request's content-type must be application/json. If any marshaling error occurs it will return
// corresponding error.
func DecodeJSONRequest(r *http.Request, dst interface{}) error {
	if r.Body == http.NoBody || r.Body == nil {
		return errdefs.InvalidParameter(errors.New("http body is required"))
	}

	if err := CheckForJSON(r); err != nil {
		return errdefs.InvalidParameter(err)
	}

	dec := json.NewDecoder(r.Body)
	// dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		var syntaxError *json.SyntaxError
		var unmarshalTypeError *json.UnmarshalTypeError

		switch {
		case errors.As(err, &syntaxError):
			return errdefs.InvalidParameter(fmt.Errorf("Request body contains badly-formed JSON (at position %d)", syntaxError.Offset))
		case errors.Is(err, io.ErrUnexpectedEOF):
			return errdefs.InvalidParameter(errors.New("Request body contains badly-formed JSON"))
		case errors.As(err, &unmarshalTypeError):
			return errdefs.InvalidParameter(fmt.Errorf("Request body contains an invalid value for the %q field (at position %d)", unmarshalTypeError.Field, unmarshalTypeError.Offset))
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return errdefs.InvalidParameter(fmt.Errorf("Request body contains unknown field: %s", fieldName))
		case errors.Is(err, io.EOF):
			return errdefs.InvalidParameter(errors.New("Request body must not be empty"))
		case err.Error() == "http: request body too large":
			return errdefs.InvalidParameter(errors.New("Request body too large"))
		default:
			return errdefs.Unknown(err)
		}
	}

	return nil
}

// DecodeRequest tries to decode given request's body into string.
func DecodeRequestString(r *http.Request) (string, error) {
	if r.Body == http.NoBody || r.Body == nil {
		return "", errdefs.InvalidParameter(errors.New("http body is required"))
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", errdefs.InvalidParameter(fmt.Errorf("failed to decode request body: %w", err))
	}

	return string(b), nil
}
