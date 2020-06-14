package errors

import (
	"errors"
	"fmt"
	"testing"
)

func TestUserMsg(t *testing.T) {
	cases := []struct {
		name string
		err  error
		want string
	}{
		{"NilErr", nil, ""},
		{"DiffErr", errors.New("stoinks"), ""},
		{"Simple", E(WithUserMsg("Deal with it!")), "Deal with it!"},
		{"Nested", E(WithErr(&Error{UserMsg: "Deal with it!"})), "Deal with it!"},
		{"NestedInAlien", fmt.Errorf("nested: %w", E(WithUserMsg("Deal with it!"))), "Deal with it!"},
		{"FirstOfMultiple", E(WithUserMsg("Bring it on!"), WithErr(E(WithUserMsg("Deal with it!")))), "Bring it on!"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := UserMsg(tc.err); got != tc.want {
				t.Errorf("UserMsg()=%q; want %q", got, tc.want)
			}
		})
	}
}
