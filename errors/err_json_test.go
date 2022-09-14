package errors

import (
	"reflect"
	"testing"
)

func TestErrorJSON(t *testing.T) {
	var toJSON JSONFunc = func(e *Error) interface{} {
		return map[string]interface{}{"op": e.Op, "kind": e.Kind}
	}
	cases := []struct {
		name string
		e    *Error
		want interface{}
	}{
		{name: "Empty", e: &Error{}, want: map[string]interface{}{"code": "", "error": "unknown error", "msg": ""}},
		{
			name: "WithFields",
			e:    E(InvalidInput, WithUserMsg("Deal with it!")).(*Error),
			want: map[string]interface{}{
				"code":  "INVALID_INPUT",
				"error": "invalid input",
				"msg":   "Deal with it!",
			},
		},
		{
			name: "WithToJSON",
			e:    E(WithOp("Get"), NotFound, WithToJSON(toJSON)).(*Error),
			want: map[string]interface{}{"op": "Get", "kind": NotFound},
		},
		{
			name: "WithNestedToJSON",
			e: E(WithOp("Get"), NotFound, WithToJSON(toJSON), WithErr(E(
				WithToJSON(func(*Error) interface{} { return 420 }),
			))).(*Error),
			want: map[string]interface{}{"op": "Get", "kind": NotFound},
		},
		{
			name: "WithOnlyNestedToJSON",
			e:    E(WithOp("Get"), NotFound, WithErr(E(WithToJSON(toJSON)))).(*Error),
			want: map[string]interface{}{"op": "Get", "kind": NotFound},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.e.JSON(); !reflect.DeepEqual(got, tc.want) {
				t.Errorf("Error.JSON()=%v; want %v", got, tc.want)
			}
		})
	}
}
