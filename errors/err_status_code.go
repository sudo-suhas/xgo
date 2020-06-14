package errors

import (
	"errors"
	"net/http"
)

// StatusCoder is implemented by any value that has a StatusCode method.
// The method is used to map the type or classification of error to the
// HTTP status code for the response from server.
type StatusCoder interface {
	StatusCode() int
}

// StatusCode attempts to determine the HTTP status code which is
// suitable for the error response. If the error does not implement
// StatusCoder or if it lacks type/classification info,
// http.StatusInternalServerError is returned. This applies for `nil`
// error as well and this case should be guarded with a nil check at the
// caller side.
func StatusCode(err error) int {
	if err == nil {
		return http.StatusInternalServerError
	}
	if e, ok := err.(StatusCoder); ok && e.StatusCode() != 0 {
		return e.StatusCode()
	}

	return StatusCode(errors.Unwrap(err))
}

// StatusCode attempts to determine the HTTP status code which is
// suitable for the error Kind.
func (e *Error) StatusCode() int {
	if e.Kind != Unknown {
		return e.Kind.Status
	}
	// Return 0 because when the error kind is not present(Unknown), try and
	// delegate to the errors down the chain.
	return 0
}
