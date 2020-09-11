package errors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"
)

func TestStatusCode(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want int
	}{
		{"NilErr", nil, http.StatusInternalServerError},
		{"DiffErr", errors.New("stoinks"), http.StatusInternalServerError},
		{"KindUnknown", E(Unknown), http.StatusInternalServerError},
		{"KindInvalidInput", E(InvalidInput), http.StatusBadRequest},
		{"KindUnauthenticated", E(Unauthenticated), http.StatusUnauthorized},
		{"KindPermissionDenied", E(PermissionDenied), http.StatusForbidden},
		{"KindNotFound", E(NotFound), http.StatusNotFound},
		{"KindConflict", E(Conflict), http.StatusConflict},
		{"KindFailedPrecondition", E(FailedPrecondition), http.StatusPreconditionFailed},
		{"KindResourceExhausted", E(ResourceExhausted), http.StatusTooManyRequests},
		{"KindInternal", E(Internal), http.StatusInternalServerError},
		{"KindCanceled", E(Canceled), http.StatusInternalServerError},
		{"KindUnimplemented", E(Unimplemented), http.StatusNotImplemented},
		{"KindUnavailable", E(Unavailable), http.StatusServiceUnavailable},
		{"KindDeadlineExceeded", E(DeadlineExceeded), http.StatusRequestTimeout},
		{"CustomErr", MyError{}, http.StatusForbidden},
		{"Nested", E(WithErr(E(InvalidInput))), http.StatusBadRequest},
		{"NestedInAlien", fmt.Errorf("nested: %w", E(InvalidInput)), http.StatusBadRequest},
		{"NestedCustom", E(WithErr(MyError{})), http.StatusForbidden},
		{"FirstOfMultiple", E(FailedPrecondition, WithErr(E(InvalidInput))), http.StatusPreconditionFailed},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := StatusCode(tc.err); got != tc.want {
				t.Errorf("StatusCode()=%d; want %d", got, tc.want)
			}
		})
	}
}
