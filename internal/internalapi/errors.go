package internalapi

import (
	"fmt"
	"net/http"
)

const (
	errorTypeInvalidRequest      = "invalid_request"
	errorTypeInternalServerError = "server_error"
	errorTypeNotFoundError       = "not_found"

	ErrorLogLevelInfo    = "info"
	ErrorLogLevelWarning = "warning"
	ErrorLogLevelError   = "error"
)

type Error struct {
	statusCode  int
	typ         string
	description string
	err         error
	logLevel    string
}

func (e *Error) StatusCode() int {
	return e.statusCode
}

func (e *Error) Type() string {
	return e.typ
}

func (e *Error) Description() string {
	return e.description
}

func (e *Error) Error() string {
	if e.err == nil {
		return e.description
	}

	return fmt.Sprintf("%s, err: %s", e.description, e.err)
}

func (e *Error) Unwrap() error {
	return e.err
}

func (e *Error) LogLevel() string {
	return e.logLevel
}

func NewInvalidRequestError(text string, err error) *Error {
	return &Error{
		statusCode:  http.StatusBadRequest,
		description: text,
		typ:         errorTypeInvalidRequest,
		err:         err,
		logLevel:    ErrorLogLevelInfo,
	}
}

func NewInvalidRequestErrorInvalidParameter(param string, err error) *Error {
	return NewInvalidRequestError(fmt.Sprintf("invalid request parameter %q", param), err)
}

func NewInvalidRequestErrorMissingRequiredParameter(param string) *Error {
	return NewInvalidRequestError(fmt.Sprintf("required parameter %q is missing", param), nil)
}

func NewServerError(err error) *Error {
	return &Error{
		statusCode:  http.StatusInternalServerError,
		description: "internal server error",
		typ:         errorTypeInternalServerError,
		err:         err,
		logLevel:    ErrorLogLevelError,
	}
}

func NewEntityNotFoundError(err error) *Error {
	return &Error{
		statusCode:  http.StatusNotFound,
		description: "entity not found",
		typ:         errorTypeNotFoundError,
		err:         err,
		logLevel:    ErrorLogLevelInfo,
	}
}
