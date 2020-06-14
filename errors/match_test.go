package errors

import (
	"database/sql"
	"io"
	"testing"
)

func TestMatch(t *testing.T) {
	msg := "Oops! Something went wrong. Please try again after some time."
	opts := Options(
		WithOp("Get"),
		WithUserMsg(msg),
		WithText("beat dead horse: already dead"),
		NotFound,
		WithData(420),
		WithErr(sql.ErrNoRows),
	)
	cases := []struct {
		err1, err2 error
		want       bool
	}{
		// Errors not of type *Error fail outright.
		{nil, nil, false},
		{sql.ErrNoRows, sql.ErrNoRows, false},
		{sql.ErrNoRows, E(WithErr(sql.ErrNoRows)), false},
		{E(WithErr(sql.ErrNoRows)), sql.ErrNoRows, false},
		// Success. We can drop fields from the first argument and still match.
		{E(WithErr(sql.ErrNoRows)), E(WithErr(sql.ErrNoRows)), true},
		{E(opts), E(opts), true},
		{E(WithOp("Get"), WithUserMsg(msg), NotFound), E(opts), true},
		{E(WithOp("Get"), WithUserMsg(msg)), E(opts), true},
		{E(WithOp("Get")), E(opts), true},
		// Failure.
		{E(WithErr(io.EOF)), E(WithErr(sql.ErrNoRows)), false},
		{E(WithOp("Op1")), E(WithOp("Op2")), false},
		{E(InvalidInput), E(PermissionDenied), false},
		{E(WithText("zoinks")), E(WithText("stoinks")), false},
		{E(WithUserMsg("deal with it")), E(WithUserMsg(msg)), false},
		{E(opts), E(opts, Internal), false},
		{E(opts), E(opts, WithText("that went as well as I expected it to")), false},
		{E(opts), E(opts, WithData(399)), false},
		{E(WithErr(sql.ErrNoRows)), E(WithOp("Op1")), false}, // Test nil error on rhs.
		// Nested *Errors.
		{
			E(WithOp("Op1"), WithErr(E(WithOp("Op2")))),
			E(WithOp("Op1"), NotFound, WithErr(E(WithOp("Op2"), WithUserMsg(msg), WithData(420)))),
			true,
		},
		{
			E(WithOp("Op1"), NotFound),
			E(WithOp("Op1"), WithUserMsg(msg), WithErr(E(WithOp("Op2"), NotFound))),
			true,
		},
	}
	for _, tc := range cases {
		if got := Match(tc.err1, tc.err2); got != tc.want {
			t.Errorf("Match(%q, %q)=%t; want %t", tc.err1, tc.err2, got, tc.want)
		}
	}
}
