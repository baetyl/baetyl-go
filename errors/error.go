package errors

import "github.com/pkg/errors"

func New(code, message string) error {
	return &CodeError{errors.New(message), code}
}

func Wrap(err error, code, message string) error {
	return &CodeError{errors.Wrap(err, message), code}
}

type CodeError struct {
	e error
	c string
}

func (e *CodeError) Code() string {
	return e.c
}

func (e *CodeError) Error() string {
	return e.e.Error()
}

func (e *CodeError) Unwrap() error {
	return e.e
}
