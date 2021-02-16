package chorderr

import (
	"fmt"
	"github.com/pkg/errors"
)

var (
	ErrDataNotExist      = BaseError{-1, "data do not exist"}
	ErrInvalidConfig     = BaseError{-2, "config is empty or invalid"}
	ErrTransportShutdown = BaseError{-3, "tcp transport is shutdown"}
	ErrNilPool            = BaseError{-4, "pool is empty"}
)

type BaseError struct {
	// error code, we usually use negative number
	code int32

	// error message
	msg string
}

func (e BaseError) Error() string {
	return fmt.Sprintf("[%d][%s]", e.code, e.msg)
}

func GetFields(e error) (code int32, msg string) {
	ee, ok := errors.Cause(e).(BaseError)
	if !ok {
		return 10000, e.Error()
	}
	return ee.code, e.Error()
}
