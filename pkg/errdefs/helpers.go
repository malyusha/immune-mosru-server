package errdefs

import (
	"context"
	"encoding/json"
)

type errNotFound struct{ error }

func (errNotFound) NotFound() {}

func (e errNotFound) Cause() error {
	return e.error
}

// NotFound is a helper to create an error of the class with the same name from any error type
func NotFound(err error) error {
	if err == nil || IsNotFound(err) {
		return err
	}
	return errNotFound{err}
}

type errInvalidParameter struct{ error }

func (errInvalidParameter) InvalidParameter() {}

func (e errInvalidParameter) Cause() error {
	return e.error
}

// InvalidParameter is a helper to create an error of the class with the same name from any error type
func InvalidParameter(err error) error {
	if err == nil || IsInvalidParameter(err) {
		return err
	}
	return errInvalidParameter{err}
}

type errUnavailable struct{ error }

func (errUnavailable) Unavailable() {}

func (e errUnavailable) Cause() error {
	return e.error
}

// Unavailable is a helper to create an error of the class with the same name from any error type
func Unavailable(err error) error {
	if err == nil || IsUnavailable(err) {
		return err
	}
	return errUnavailable{err}
}

type errSystem struct{ error }

func (errSystem) System() {}

func (e errSystem) Cause() error {
	return e.error
}

// System is a helper to create an error of the class with the same name from any error type
func System(err error) error {
	if err == nil || IsSystem(err) {
		return err
	}
	return errSystem{err}
}

type errNotImplemented struct{ error }

func (errNotImplemented) NotImplemented() {}

func (e errNotImplemented) Cause() error {
	return e.error
}

// NotImplemented is a helper to create an error of the class with the same name from any error type
func NotImplemented(err error) error {
	if err == nil || IsNotImplemented(err) {
		return err
	}
	return errNotImplemented{err}
}

type errUnknown struct{ error }

func (errUnknown) Unknown() {}

func (e errUnknown) Cause() error {
	return e.error
}

// Unknown is a helper to create an error of the class with the same name from any error type
func Unknown(err error) error {
	if err == nil || IsUnknown(err) {
		return err
	}
	return errUnknown{err}
}

type errCancelled struct{ error }

func (errCancelled) Cancelled() {}

func (e errCancelled) Cause() error {
	return e.error
}

// Cancelled is a helper to create an error of the class with the same name from any error type
func Cancelled(err error) error {
	if err == nil || IsCancelled(err) {
		return err
	}
	return errCancelled{err}
}

type errDeadline struct{ error }

func (errDeadline) DeadlineExceeded() {}

func (e errDeadline) Cause() error {
	return e.error
}

// Deadline is a helper to create an error of the class with the same name from any error type
func Deadline(err error) error {
	if err == nil || IsDeadline(err) {
		return err
	}
	return errDeadline{err}
}

type errValidation struct {
	error
	fields InvalidFields
}

func (e *errValidation) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Message string        `json:"message"`
		Fields  InvalidFields `json:"fields"`
	}{
		Message: e.Error(),
		Fields:  e.fields,
	})
}

func (e errValidation) Fields() InvalidFields {
	return e.fields
}
func (e errValidation) Cause() error {
	return e.error
}

// Deadline is a helper to create an error of the class with the same name from any error type
func Validation(err error) error {
	if err == nil || IsValidation(err) {
		return err
	}

	return errValidation{error: err, fields: make(InvalidFields)}
}

// FromContext returns the error class from the passed in context
func FromContext(ctx context.Context) error {
	e := ctx.Err()
	if e == nil {
		return nil
	}

	if e == context.Canceled {
		return Cancelled(e)
	}
	if e == context.DeadlineExceeded {
		return Deadline(e)
	}
	return Unknown(e)
}
