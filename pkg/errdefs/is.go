package errdefs

type causer interface {
	Cause() error
}

func getImplementer(err error) error {
	switch e := err.(type) {
	case
			ErrNotFound,
			ErrInvalidParameter,
			ErrValidation,
			ErrUnavailable,
			ErrSystem,
			ErrNotImplemented,
			ErrCancelled,
			ErrDeadline,
			ErrUnknown:
		return err
	case causer:
		return getImplementer(e.Cause())
	default:
		return err
	}
}

// IsNotFound returns if the passed in error is an ErrNotFound
func IsNotFound(err error) bool {
	_, ok := getImplementer(err).(ErrNotFound)
	return ok
}

// IsInvalidParameter returns if the passed in error is an ErrInvalidParameter
func IsInvalidParameter(err error) bool {
	_, ok := getImplementer(err).(ErrInvalidParameter)
	return ok
}

func IsUnavailable(err error) bool {
	_, ok := getImplementer(err).(ErrUnavailable)
	return ok
}

// IsSystem returns if the passed in error is an ErrSystem
func IsSystem(err error) bool {
	_, ok := getImplementer(err).(ErrSystem)
	return ok
}

// IsNotImplemented returns if the passed in error is an ErrNotImplemented
func IsNotImplemented(err error) bool {
	_, ok := getImplementer(err).(ErrNotImplemented)
	return ok
}

// IsUnknown returns if the passed in error is an ErrUnknown
func IsUnknown(err error) bool {
	_, ok := getImplementer(err).(ErrUnknown)
	return ok
}

// IsValidation returns if the passed in error is an ErrValidation
func IsValidation(err error) bool {
	_, ok := getImplementer(err).(ErrValidation)
	return ok
}

// IsCancelled returns if the passed in error is an ErrCancelled
func IsCancelled(err error) bool {
	_, ok := getImplementer(err).(ErrCancelled)
	return ok
}

// IsDeadline returns if the passed in error is an ErrDeadline
func IsDeadline(err error) bool {
	_, ok := getImplementer(err).(ErrDeadline)
	return ok
}
