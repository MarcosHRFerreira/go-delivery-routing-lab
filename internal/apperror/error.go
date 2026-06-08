// Package apperror provides application-level errors with HTTP semantics.
package apperror

import "net/http"

type Error struct {
	statusCode int
	message    string
	cause      error
}

func New(statusCode int, message string) *Error {
	return &Error{
		statusCode: statusCode,
		message:    message,
	}
}

func Wrap(statusCode int, message string, cause error) *Error {
	return &Error{
		statusCode: statusCode,
		message:    message,
		cause:      cause,
	}
}

func Internal(message string, cause error) *Error {
	return Wrap(http.StatusInternalServerError, message, cause)
}

func BadRequest(message string) *Error {
	return New(http.StatusBadRequest, message)
}

func Unauthorized(message string) *Error {
	return New(http.StatusUnauthorized, message)
}

func Forbidden(message string) *Error {
	return New(http.StatusForbidden, message)
}

func NotFound(message string) *Error {
	return New(http.StatusNotFound, message)
}

func Conflict(message string) *Error {
	return New(http.StatusConflict, message)
}

func (e *Error) Error() string {
	return e.message
}

func (e *Error) StatusCode() int {
	return e.statusCode
}

func (e *Error) Unwrap() error {
	return e.cause
}

func StatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}

	if asStatusCode, ok := err.(interface{ StatusCode() int }); ok {
		return asStatusCode.StatusCode()
	}

	return http.StatusInternalServerError
}
