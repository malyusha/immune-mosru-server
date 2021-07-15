package telegram

import (
	"errors"
	"fmt"
)

type internalError struct {
	err error
}

func (i *internalError) Error() string {
	return fmt.Sprintf("internal error: %s", i.err.Error())
}

func newInternalError(err error) *internalError {
	if err == nil || IsInternal(err) {
		return err.(*internalError)
	}

	return &internalError{
		err: err,
	}
}

func IsInternal(err error) bool {
	var iErr *internalError
	return errors.As(err, &iErr)
}
