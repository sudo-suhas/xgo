package errors

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/sudo-suhas/xgo"
)

func TestErrorDetails(t *testing.T) {
	cases := []struct {
		name string
		e    *Error
		want InternalDetails
	}{
		{name: "Empty", e: &Error{}, want: InternalDetails{Error: "no error"}},
		{
			name: "WithFields",
			e: E(
				WithOp("Get"),
				WithUserMsg("irrelevant"),
				WithText("beat dead horse: already dead"),
				NotFound,
				WithData(420),
				WithErr(sql.ErrNoRows),
			).(*Error),
			want: InternalDetails{
				Ops:   []xgo.Op{"Get"},
				Kind:  NotFound,
				Error: "Get: not found: beat dead horse: already dead: sql: no rows in result set",
				Data:  420,
			},
		},
		{
			name: "WithFieldsNested",
			e: E(
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
			).(*Error),
			want: InternalDetails{
				Ops:   []xgo.Op{"Get", "Select"},
				Kind:  Internal,
				Error: "Get: internal error: beat dead horse: already dead: Select: not found: select data from table where column = ?: sql: no rows in result set",
				Data:  []interface{}{420, "xyz"},
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.e.Details(); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("\nError.Details()=%#v; \nwant %#v", got, tc.want)
			}
		})
	}
}
