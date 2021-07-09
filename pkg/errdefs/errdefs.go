package errdefs

// ErrNotFound signals that the requested object doesn't exist
type ErrNotFound interface {
	NotFound()
}

// ErrInvalidParameter signals that the user input is invalid
type ErrInvalidParameter interface {
	InvalidParameter()
}

// ErrUnavailable signals that the requested action/subsystem is not available.
type ErrUnavailable interface {
	Unavailable()
}

// ErrSystem signals that some internal error occurred.
type ErrSystem interface {
	System()
}

// ErrNotImplemented signals that the requested action/feature is not implemented on the system as configured.
type ErrNotImplemented interface {
	NotImplemented()
}

// InvalidFields represents struct, holding fields and their list of error messages.
type InvalidFields map[string][]string

// Add adds message to existing field or creates new field with list of errors if not initialized.
func (f InvalidFields) Add(field, message string) {
	if _, ok := f[field]; ok {
		f[field] = append(f[field], message)
	} else {
		f[field] = []string{message}
	}
}

// ErrValidation represents validation errors inside request.
type ErrValidation interface {
	Fields() InvalidFields
}

// ErrUnknown signals that the kind of error that occurred is not known.
type ErrUnknown interface {
	Unknown()
}

// ErrCancelled signals that the action was cancelled.
type ErrCancelled interface {
	Cancelled()
}

// ErrDeadline signals that the deadline was reached before the action completed.
type ErrDeadline interface {
	DeadlineExceeded()
}
