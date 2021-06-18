package response

import (
	"fmt"
)

// HTTPError is a error type support for http request
type HTTPError interface {
	Error() string
	Unwarp() error
	Code() int // http status code
	String() string
	Message() string
}

type httpError struct {
	code    int // http status code
	message string
	err     error
}

// NewError create new http error
func NewError(code int, msg string, err error) HTTPError {
	if err == nil {
		panic("create new error but got none error")
	}
	httpError{}.Code()
	return httpError{
		code:    code,
		message: msg,
		err:     err,
	}
}

func (e httpError) Error() string {
	return fmt.Sprintf("message: %s, error: %s", e.message, e.err.Error())
}

func (e httpError) Unwarp() error {
	return e.err
}

// Code get http status code
func (e httpError) Code() int {
	return e.code
}

func (e httpError) String() string {
	return fmt.Sprintf("get http error, code: %d, message: %s, error: %s", e.code, e.message, e.err.Error())
}

func (e httpError) Message() string {
	return e.message
}
