package httputils

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// WriteJSON writes the value v to the http response stream as json with standard json encoding.
func WriteJSON(w http.ResponseWriter, code int, v interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)

	err := enc.Encode(v)
	if err != nil {
		return fmt.Errorf(fmt.Sprintf("failed to encode response with %d status code: %s", code, err))
	}

	return nil
}

func JSONOk(w http.ResponseWriter, v interface{}) error {
	return WriteJSON(w, http.StatusOK, v)
}

// WriteResponse detects content type of request and responds with corresponding data and content-type.
func WriteResponse(w http.ResponseWriter, r *http.Request, code int, v interface{}) error {
	ct := r.Header.Get("Accept")
	if ct == "" || matchesContentType("application/json", ct) {
		return WriteJSON(w, code, v)
	}

	s, ok := v.(string)
	if !ok {
		return errors.New("string expected to response with text")
	}

	return WriteText(w, code, s)
}

func WriteText(w http.ResponseWriter, code int, v string) error {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(code)
	if _, err := w.Write([]byte(v)); err != nil {
		return err
	}

	return nil
}
