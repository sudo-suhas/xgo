package errors

import (
	"errors"
	"net/http"
	"strings"
)

// Kind is the type of error. It is the tuple of the error code and HTTP
// status code. Defining custom Kinds in application domain is
// recommended if the predeclared Kinds are not suitable.
type Kind struct {
	Code   string
	Status int
}

// Apply implements Option interface. This allows Kind to be passes as a
// constructor option.
func (k Kind) Apply(e *Error) {
	e.Kind = k
}

// GetKind is implemented by any value that has a GetKind method. The
// method is used to determine the type or classification of error.
type GetKind interface {
	GetKind() Kind
}

// WhatKind returns the Kind associated with the given error. If the
// error is nil or does not implement GetKind interface, Unknown is
// returned.
func WhatKind(err error) Kind {
	if err == nil {
		return Unknown
	}

	if e, ok := err.(GetKind); ok && e.GetKind() != Unknown {
		return e.GetKind()
	}

	return WhatKind(errors.Unwrap(err))
}

// Error kinds are adapted from
// https://github.com/grpc/grpc-go/blob/v1.12.0/codes/codes.go

// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Status#Client_error_responses

var (
	// Unknown error. An example of where this error may be returned is
	// if a Status value received from another address space belongs to
	// an error-space that is not known in this address space. Also
	// errors raised by APIs that do not return enough error information
	// may be converted to this error.
	Unknown = Kind{}

	// Canceled indicates the operation was canceled (typically by the caller).
	Canceled = Kind{
		Code:   "CANCELED",
		Status: http.StatusInternalServerError, // 500
	}

	// InvalidInput indicates client specified an invalid input.
	// Note that this differs from FailedPrecondition. It indicates arguments
	// that are problematic regardless of the state of the system
	// (e.g., a malformed file name).
	InvalidInput = Kind{
		Code:   "INVALID_INPUT",
		Status: http.StatusBadRequest, // 400
	}

	// DeadlineExceeded means operation expired before completion.
	// For operations that change the state of the system, this error may be
	// returned even if the operation has completed successfully. For
	// example, a successful response from a server could have been delayed
	// long enough for the deadline to expire.
	DeadlineExceeded = Kind{
		Code:   "DEADLINE_EXCEEDED",
		Status: http.StatusRequestTimeout, // 408
	}

	// NotFound means some requested entity (e.g., file or directory) was
	// not found.
	NotFound = Kind{
		Code:   "NOT_FOUND",
		Status: http.StatusNotFound, // 404
	}

	// AlreadyExists means an attempt to create an entity failed because one
	// already exists.
	AlreadyExists = Kind{
		Code:   "ALREADY_EXISTS",
		Status: 0,
	}

	// PermissionDenied indicates the caller does not have permission to
	// execute the specified operation. It must not be used for rejections
	// caused by exhausting some resource (use ResourceExhausted instead
	// for those errors). It must not be used if the caller cannot be
	// identified (use Unauthenticated instead for those errors).
	PermissionDenied = Kind{
		Code:   "PERMISSION_DENIED",
		Status: http.StatusForbidden, // 403
	}

	// ResourceExhausted indicates some resource has been exhausted, perhaps
	// a per-user quota, or perhaps the entire file system is out of space.
	ResourceExhausted = Kind{
		Code:   "RESOURCE_EXHAUSTED",
		Status: http.StatusTooManyRequests, // 429
	}

	// FailedPrecondition indicates operation was rejected because the
	// system is not in a state required for the operation's execution.
	// For example, directory to be deleted may be non-empty, an rmdir
	// operation is applied to a non-directory, etc.
	//
	// A litmus test that may help a service implementor in deciding
	// between FailedPrecondition, Aborted, and Unavailable:
	//  (a) Use Unavailable if the client can retry just the failing call.
	//  (b) Use Aborted if the client should retry at a higher-level
	//      (e.g., restarting a read-modify-write sequence).
	//  (c) Use FailedPrecondition if the client should not retry until
	//      the system state has been explicitly fixed. E.g., if an "rmdir"
	//      fails because the directory is non-empty, FailedPrecondition
	//      should be returned since the client should not retry unless
	//      they have first fixed up the directory by deleting files from it.
	//  (d) Use FailedPrecondition if the client performs conditional
	//      REST Get/Update/Delete on a resource and the resource on the
	//      server does not match the condition. E.g., conflicting
	//      read-modify-write on the same resource.
	FailedPrecondition = Kind{
		Code:   "FAILED_PRECONDITION",
		Status: http.StatusPreconditionFailed, // 412
	}

	// Aborted indicates the operation was aborted, typically due to a
	// concurrency issue like sequencer check failures, transaction aborts,
	// etc.
	//
	// See litmus test above for deciding between FailedPrecondition,
	// Aborted, and Unavailable.
	Aborted = Kind{
		Code:   "ABORTED",
		Status: 0,
	}

	// Unimplemented indicates operation is not implemented or not
	// supported/enabled in this service.
	Unimplemented = Kind{
		Code:   "UNIMPLEMENTED",
		Status: http.StatusNotImplemented, // 501
	}

	// Internal errors. Means some invariants expected by underlying
	// system has been broken. If you see one of these errors,
	// something is very broken.
	Internal = Kind{
		Code:   "INTERNAL",
		Status: http.StatusInternalServerError, // 500
	}

	// Unavailable indicates the service is currently unavailable.
	// This is a most likely a transient condition and may be corrected
	// by retrying with a backoff. Note that it is not always safe to retry
	// non-idempotent operations.
	//
	// See litmus test above for deciding between FailedPrecondition,
	// Aborted, and Unavailable.
	Unavailable = Kind{
		Code:   "UNAVAILABLE",
		Status: http.StatusServiceUnavailable, // 503
	}

	// Unauthenticated indicates the request does not have valid
	// authentication credentials for the operation.
	Unauthenticated = Kind{
		Code:   "UNAUTHENTICATED",
		Status: http.StatusUnauthorized, // 401
	}

	// IO indicates external I/O error such as network failure.
	IO = Kind{
		Code:   "IO",
		Status: 0,
	}

	// Illegal indicates an action which is illegal and not allowed.
	Illegal = Kind{
		Code:   "ILLEGAL",
		Status: 0,
	}
)

// KindFromStatus returns the Kind based on the given HTTP status code.
//
// It is not aware of any Kind defined in the application domain.
func KindFromStatus(status int) Kind {
	switch status {
	case http.StatusBadRequest, http.StatusUnprocessableEntity:
		return InvalidInput

	case http.StatusUnauthorized:
		return Unauthenticated

	case http.StatusForbidden:
		return PermissionDenied

	case http.StatusNotFound:
		return NotFound

	case http.StatusRequestTimeout:
		return DeadlineExceeded

	case http.StatusConflict:
		// TODO: ??

	case http.StatusPreconditionFailed:
		return FailedPrecondition

	case http.StatusTooManyRequests:
		return ResourceExhausted

	case http.StatusInternalServerError:
		return Internal

	case http.StatusNotImplemented:
		return Unimplemented

	case http.StatusServiceUnavailable:
		return Unavailable
	}
	return Unknown
}

// KindFromCode returns the error kind based on the given error code.
//
// It is not aware of any Kind defined in the application domain.
func KindFromCode(code string) Kind {
	switch code {
	case "CANCELED":
		return Canceled

	case "INVALID_INPUT":
		return InvalidInput

	case "DEADLINE_EXCEEDED":
		return DeadlineExceeded

	case "NOT_FOUND":
		return NotFound

	case "ALREADY_EXISTS":
		return AlreadyExists

	case "PERMISSION_DENIED":
		return PermissionDenied

	case "RESOURCE_EXHAUSTED":
		return ResourceExhausted

	case "FAILED_PRECONDITION":
		return FailedPrecondition

	case "ABORTED":
		return Aborted

	case "UNIMPLEMENTED":
		return Unimplemented

	case "INTERNAL":
		return Internal

	case "UNAVAILABLE":
		return Unavailable

	case "UNAUTHENTICATED":
		return Unauthenticated

	case "IO":
		return IO

	case "ILLEGAL":
		return Illegal

	}
	return Unknown
}

func (k Kind) String() string {
	switch k {
	case IO:
		return "I/O error"

	case Unknown:
		return "unknown error"

	case Illegal:
		return "illegal action, not allowed"

	case Internal:
		return "internal error"
	}
	return strings.ToLower(
		strings.Join(
			strings.Split(k.Code, "_"),
			" ",
		),
	)
}
