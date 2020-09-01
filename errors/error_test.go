package errors

import (
	"database/sql"
	"errors"
	"reflect"
	"testing"

	"github.com/sudo-suhas/xgo"
)

// Compile time check to ensure type implements the interfaces.
var (
	_ GetKind     = (*Error)(nil)
	_ StatusCoder = (*Error)(nil)
	_ xgo.JSONer  = (*Error)(nil)
)

func TestE(t *testing.T) {
	cases := []struct{ got, want error }{
		// Options are being applied
		{E(WithOp("Get")), &Error{Op: "Get"}},
		{
			E(WithOp("Get"), NotFound, WithTextf("boom %d", 5)),
			&Error{Op: "Get", Kind: NotFound, Text: "boom 5"},
		},
		// Among options targeting the same field, the latter one wins.
		{
			E(WithOp("First"), NotFound, WithOp("Second")),
			&Error{Op: "Second", Kind: NotFound},
		},
		// Chaining errors of type other than *Error
		{E(WithErr(errors.New("stoinks"))), &Error{Err: errors.New("stoinks")}},
		{E(WithErr(MyError{})), &Error{Err: MyError{}}},
		// De-duplication
		{E(WithOp("Get"), WithErr(E(WithOp("Get")))), &Error{Op: "Get"}},
		{E(InvalidInput, WithErr(E(InvalidInput))), &Error{Kind: InvalidInput}},
		{
			E(WithUserMsg("Deal with it!"), WithErr(E(WithUserMsg("Deal with it!")))),
			&Error{UserMsg: "Deal with it!"},
		},
		{
			E(WithText("beat dead horse: already dead"), WithErr(E(WithText("beat dead horse: already dead")))),
			&Error{Text: "beat dead horse: already dead"},
		},
		// Lift field from child error
		{E(WithErr(E(WithOp("Get")))), &Error{Op: "Get"}},
		{E(WithErr(E(InvalidInput))), &Error{Kind: InvalidInput}},
		{E(WithErr(E(WithUserMsg("Deal with it!")))), &Error{UserMsg: "Deal with it!"}},
		{E(WithErr(E(WithData(420)))), &Error{Data: 420}},
		{
			E(WithErr(E(WithText("beat dead horse: already dead")))),
			&Error{Text: "beat dead horse: already dead"},
		},
		{E(WithErr(E(WithErr(sql.ErrNoRows)))), &Error{Err: sql.ErrNoRows}},
		// Child Err with Op/Kind, cannot lift Text, Err
		{
			E(Internal, WithErr(E(NotFound, WithText("beat dead horse: already dead")))),
			&Error{Kind: Internal, Err: &Error{Kind: NotFound, Text: "beat dead horse: already dead"}},
		},
		{
			E(Internal, WithErr(E(NotFound, WithErr(sql.ErrNoRows)))),
			&Error{Kind: Internal, Err: &Error{Kind: NotFound, Err: sql.ErrNoRows}},
		},
		{
			E(WithOp("Op1"), WithErr(E(WithOp("Op2"), WithText("beat dead horse: already dead")))),
			&Error{Op: "Op1", Err: &Error{Op: "Op2", Text: "beat dead horse: already dead"}},
		},
		{
			E(WithOp("Op1"), WithErr(E(WithOp("Op2"), WithErr(sql.ErrNoRows)))),
			&Error{Op: "Op1", Err: &Error{Op: "Op2", Err: sql.ErrNoRows}},
		},
		// Cannot lift Err if anything else is present
		{
			E(WithText("text 1"), WithErr(E(WithText("text 2"), WithErr(sql.ErrNoRows)))),
			&Error{Text: "text 1", Err: &Error{Text: "text 2", Err: sql.ErrNoRows}},
		},
		{
			E(WithUserMsg("msg a"), WithErr(E(WithUserMsg("msg b"), WithErr(sql.ErrNoRows)))),
			&Error{UserMsg: "msg a", Err: &Error{UserMsg: "msg b", Err: sql.ErrNoRows}},
		},
		{
			E(WithData(1), WithErr(E(WithData(2), WithErr(sql.ErrNoRows)))),
			&Error{Data: 1, Err: &Error{Data: 2, Err: sql.ErrNoRows}},
		},
		{
			E(WithToJSON(func(*Error) interface{} { return 1 }), WithErr(E(WithToJSON(func(*Error) interface{} { return 2 }), WithErr(sql.ErrNoRows)))),
			&Error{ToJSON: func(*Error) interface{} { return 1 }, Err: &Error{ToJSON: func(*Error) interface{} { return 2 }, Err: sql.ErrNoRows}},
		},
		// Nested with fully loaded errors
		{
			E(
				WithOp("Get"),
				WithUserMsg("Deal with it!"),
				WithErr(E(
					WithOp("Select"),
					NotFound,
					WithText("select data from table where column = ?"),
					WithData("xyz"),
					WithErr(sql.ErrNoRows),
				)),
			),
			&Error{
				Op:      "Get",
				Kind:    NotFound,
				UserMsg: "Deal with it!",
				Data:    "xyz",
				Err: &Error{
					Op:   "Select",
					Text: "select data from table where column = ?",
					Err:  sql.ErrNoRows,
				},
			},
		},
		{
			E(
				WithOp("Get"),
				WithUserMsg("Deal with it!"),
				WithText("beat dead horse: already dead"),
				Internal,
				WithData(420),
				WithErr(E(
					WithOp("Select"),
					NotFound,
					WithText("select data from table where column = ?"),
					WithData("xyz"),
					WithErr(sql.ErrNoRows),
				)),
			),
			&Error{
				Op:      "Get",
				Kind:    Internal,
				UserMsg: "Deal with it!",
				Text:    "beat dead horse: already dead",
				Data:    420,
				Err: &Error{
					Op:   "Select",
					Kind: NotFound,
					Text: "select data from table where column = ?",
					Data: "xyz",
					Err:  sql.ErrNoRows,
				},
			},
		},
	}
	for _, tc := range cases {
		if !Match(tc.want, tc.got) {
			t.Errorf("\nE()=%#v \nwant %#v", tc.got, tc.want)
		}
	}
}

func TestDoesNotChangePreviousError(t *testing.T) {
	err := E(PermissionDenied)
	err2 := E(WithOp("I will NOT modify err"), WithErr(err))

	expected := "I will NOT modify err: permission denied"
	if err2.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err2)
	}
	if kind := err.(*Error).Kind; kind != PermissionDenied {
		t.Errorf("Expected kind %v, got %v", PermissionDenied, kind)
	}
}

func TestError(t *testing.T) {
	cases := []struct {
		err  error
		want string
	}{
		{&Error{}, "no error"},
		{E(WithOp("Get")), "Get"},
		{E(NotFound), "not found"},
		{E(WithOp("Get"), NotFound), "Get: not found"},
		{E(WithText("boom")), "boom"},
		{E(WithOp("Get"), WithText("boom")), "Get: boom"},
		{E(NotFound, WithText("boom")), "not found: boom"},
		{E(WithOp("Get"), NotFound, WithText("boom")), "Get: not found: boom"},
		{
			E(
				WithOp("Get"),
				WithUserMsg("irrelevant"),
				WithText("beat dead horse: already dead"),
				Internal,
				WithData(420),
				WithErr(E(
					WithOp("Select"),
					NotFound,
					WithText("select data from table where column = ?"),
					WithData("xyz"),
					WithErr(sql.ErrNoRows),
				)),
			),
			"Get: internal error: beat dead horse: already dead: Select: not found: select data from table where column = ?: sql: no rows in result set",
		},
	}
	for _, tc := range cases {
		if got := tc.err.Error(); got != tc.want {
			t.Errorf("Error.Error()=%q; want %q", got, tc.want)
		}
	}
}

func TestErrorOps(t *testing.T) {
	tests := []struct {
		name string
		e    *Error
		want []xgo.Op
	}{
		{"WithoutOp", E(InvalidInput, WithText("divide by zero")).(*Error), nil},
		{"WithOp", E(WithOp("Get")).(*Error), []xgo.Op{"Get"}},
		{
			"WithOps",
			E(
				WithOp("Op1"),
				WithErr(E(
					WithOp("Op2"),
					WithErr(&Error{Err: E(WithOp("Op3"))}),
				)),
			).(*Error),
			[]xgo.Op{"Op1", "Op2", "Op3"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Ops(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Error.Ops()=%v; want %v", got, tt.want)
			}
		})
	}
}
