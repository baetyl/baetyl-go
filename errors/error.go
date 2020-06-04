package errors

import (
	"fmt"
	"github.com/pkg/errors"
)

type Coder interface {
	Code() string
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

type codeError struct {
	e error
	c string
}

func New(code, message string) error {
	return &codeError{errors.New(message), code}
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
