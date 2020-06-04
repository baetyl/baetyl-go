package errors

import (
	"fmt"

	"github.com/pkg/errors"
)

type Coder interface {
	Code() string
}

func New(message string) error {
	return errors.New(message)
}

func Errorf(format string, args ...interface{}) error {
	return errors.Errorf(format, args...)
}

func Cause(err error) error {
	return errors.Cause(err)
}

func Trace(err error) error {
	if err == nil {
		return nil
	}
	switch err.(type) {
	case fmt.Formatter:
		return err
	default:
		return errors.WithStack(err)
	}
}

func CodeError(code, message string) error {
	return &codeError{errors.New(message), code}
}

type codeError struct {
	e error
	c string
}

func (e *codeError) Code() string {
	return e.c
}

func (e *codeError) Error() string {
	return e.e.Error()
}

func (e *codeError) Format(s fmt.State, verb rune) {
	e.e.(fmt.Formatter).Format(s, verb)
}
