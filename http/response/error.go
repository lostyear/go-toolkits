package response

import (
	"errors"
	"fmt"
	"net/http"
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

var (
	// ErrServerError is error of http internal server error
	ErrServerError = errors.New("Internal Server Error")
	// ErrNotFound    is error of http not found
	ErrNotFound = errors.New("Not Found")
	// ErrForbidden   is error of http forbidden
	ErrForbidden = errors.New("Forbidden")
	// ErrBadRequest  is error of http bad request
	ErrBadRequest = errors.New("Bad Request")
)

// NewSimpleError create a response which show internal server error
func NewSimpleError(err error) HTTPError {
	return httpError{
		code:    http.StatusInternalServerError,
		message: err.Error(),
		err:     err,
	}
}

// NewServerError create a response which show internal server error
func NewServerError(msg string) HTTPError {
	return httpError{
		code:    http.StatusInternalServerError,
		message: msg,
		err:     ErrServerError,
	}
}

// NewNotFoundError create a response which show not found
func NewNotFoundError(msg string) HTTPError {
	return httpError{
		code:    http.StatusNotFound,
		message: msg,
		err:     ErrNotFound,
	}
}

// NewForbiddenError create a response which show forbidden
func NewForbiddenError(msg string) HTTPError {
	return httpError{
		code:    http.StatusForbidden,
		message: msg,
		err:     ErrForbidden,
	}
}

// NewBadRequestError create a response which show bad request
func NewBadRequestError(msg string) HTTPError {
	return httpError{
		code:    http.StatusBadRequest,
		message: msg,
		err:     ErrBadRequest,
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
