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
		{"KindCanceled", E(Canceled), http.StatusInternalServerError},
		{"KindInvalidInput", E(InvalidInput), http.StatusBadRequest},
		{"KindDeadlineExceeded", E(DeadlineExceeded), http.StatusRequestTimeout},
		{"KindNotFound", E(NotFound), http.StatusNotFound},
		{"KindAlreadyExists", E(AlreadyExists), http.StatusInternalServerError},
		{"KindPermissionDenied", E(PermissionDenied), http.StatusForbidden},
		{"KindResourceExhausted", E(ResourceExhausted), http.StatusTooManyRequests},
		{"KindFailedPrecondition", E(FailedPrecondition), http.StatusPreconditionFailed},
		{"KindAborted", E(Aborted), http.StatusInternalServerError},
		{"KindUnimplemented", E(Unimplemented), http.StatusNotImplemented},
		{"KindInternal", E(Internal), http.StatusInternalServerError},
		{"KindUnavailable", E(Unavailable), http.StatusServiceUnavailable},
		{"KindUnauthenticated", E(Unauthenticated), http.StatusUnauthorized},
		{"KindIO", E(IO), http.StatusInternalServerError},
		{"KindIllegal", E(Illegal), http.StatusInternalServerError},
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
