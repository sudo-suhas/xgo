package errors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestWhatKind(t *testing.T) {
	var (
		customErr MyError
		inputErr  = E(InvalidInput, WithText("zoinks"))
		noKindErr = E(WithText("no kind"))
	)
	cases := []struct {
		err  error
		want Kind
	}{
		// Non-Error errors.
		{nil, Unknown},
		{errors.New("not an *Error"), Unknown},

		// Basic errors.
		{customErr, Internal},
		{inputErr, InvalidInput},
		{noKindErr, Unknown},

		// Nested *Error values.
		{E(WithText("nesting"), WithErr(customErr)), Internal},
		{E(WithText("nesting"), WithErr(inputErr)), InvalidInput},
		{E(WithText("nesting"), WithErr(noKindErr)), Unknown},
		{fmt.Errorf("nested: %w", E(InvalidInput)), InvalidInput},

		// Precedence for *Error at the top
		{E(WithText("nesting"), PermissionDenied, WithErr(customErr)), PermissionDenied},
		{E(WithText("nesting"), PermissionDenied, WithErr(inputErr)), PermissionDenied},
	}
	for _, tc := range cases {
		if got := WhatKind(tc.err); got != tc.want {
			t.Errorf("WhatKind(%q)=%q; want %q", tc.err, got, tc.want)
		}
	}
}

func TestKindFromStatus(t *testing.T) {
	cases := []struct {
		status int
		want   Kind
	}{
		{http.StatusBadRequest, InvalidInput},
		{http.StatusUnprocessableEntity, InvalidInput},
		{http.StatusUnauthorized, Unauthenticated},
		{http.StatusForbidden, PermissionDenied},
		{http.StatusNotFound, NotFound},
		{http.StatusConflict, Conflict},
		{http.StatusPreconditionFailed, FailedPrecondition},
		{http.StatusTooManyRequests, ResourceExhausted},
		{http.StatusInternalServerError, Internal},
		{http.StatusNotImplemented, Unimplemented},
		{http.StatusServiceUnavailable, Unavailable},
		{http.StatusTeapot, Unknown},
	}
	for _, tc := range cases {
		if got := KindFromStatus(tc.status); got != tc.want {
			t.Errorf("KindFromStatus(%d)=%#v; want %#v", tc.status, got, tc.want)
		}
	}
}

func TestKindFromCode(t *testing.T) {
	cases := []struct {
		code string
		want Kind
	}{
		{"CANCELED", Canceled},
		{"INVALID_INPUT", InvalidInput},
		{"DEADLINE_EXCEEDED", DeadlineExceeded},
		{"NOT_FOUND", NotFound},
		{"CONFLICT", Conflict},
		{"PERMISSION_DENIED", PermissionDenied},
		{"RESOURCE_EXHAUSTED", ResourceExhausted},
		{"FAILED_PRECONDITION", FailedPrecondition},
		{"UNIMPLEMENTED", Unimplemented},
		{"INTERNAL", Internal},
		{"UNAVAILABLE", Unavailable},
		{"UNAUTHENTICATED", Unauthenticated},
		{"UNDEFINED", Unknown},
	}
	for _, tc := range cases {
		if got := KindFromCode(tc.code); got != tc.want {
			t.Errorf("KindFromCode(%q)=%#v; want %#v", tc.code, got, tc.want)
		}
	}
}

func TestKindString(t *testing.T) {
	cases := []struct {
		kind Kind
		want string
	}{
		{Kind{}, "unknown error"},
		{Unknown, "unknown error"},
		{Internal, "internal error"},
		{Canceled, "canceled"},
		{InvalidInput, "invalid input"},
		{DeadlineExceeded, "deadline exceeded"},
		{NotFound, "not found"},
		{Conflict, "conflict"},
		{PermissionDenied, "permission denied"},
		{ResourceExhausted, "resource exhausted"},
		{FailedPrecondition, "failed precondition"},
		{Unimplemented, "unimplemented"},
		{Unavailable, "unavailable"},
		{Unauthenticated, "unauthenticated"},
	}
	for _, tc := range cases {
		if got := tc.kind.String(); got != tc.want {
			t.Errorf("%#v.String()=%q; want %q", tc.kind, got, tc.want)
		}
	}
}
