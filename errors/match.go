package errors

import (
	"fmt"
	"reflect"
)

// Match compares its two error arguments. It can be used to check for
// expected errors in tests. Both arguments must have underlying type
// *Error or Match will return false. Otherwise, it returns true iff
// every non-zero field of the template error is equal to the
// corresponding field of the error being checked. If the Err field is
// a *Error, Match recurs on that field; otherwise it compares the
// strings returned by the Error methods. Elements that are in the
// second argument but not present in the first are ignored.
//
// For example,
//	errors.Match(errors.E(errors.WithOp("service.MakeBooking"), errors.PermissionDenied), err)
// tests whether err is an Error with Kind=PermissionDenied and
// Op=service.MakeBooking.
func Match(template, err error) bool {
	return len(Diff(template, err)) == 0
}

// Diff compares and returns the diff between two error arguments. It
// can be used to extract the exact details of difference between
// expected and got errors in tests. It returns an empty array iff
// both arguments are of type *Error and every non-zero field of the
// template error is equal to the corresponding field of the second. If
// the Err field is a *Error, Diff recurs on that field; otherwise it
// compares the strings returned by the Error methods. Elements that are
// in the second argument but not present in the template are ignored.
//
// For example,
// 	errors.Diff(errors.E(errors.WithOp("service.MakeBooking"), errors.PermissionDenied), err)
// tests whether err is an Error with Kind=PermissionDenied and
// Op=service.MakeBooking.
func Diff(template, err error) []string { //nolint: gocognit
	t, ok1 := template.(*Error)
	e, ok2 := err.(*Error)
	if !ok1 || !ok2 {
		return []string{fmt.Sprintf(`template="%v"(type %[1]T); err="%v"(type %[2]T)`, template, err)}
	}

	var diff []string
	if t.Op != "" && t.Op != e.Op {
		diff = append(diff, fmt.Sprintf("Op: template=%q; err=%q", t.Op, e.Op))
	}
	if t.Kind != Unknown && t.Kind != e.Kind {
		diff = append(diff, fmt.Sprintf("Kind: template=%q; err=%q", t.Kind, e.Kind))
	}
	if t.Text != "" && t.Text != e.Text {
		diff = append(diff, fmt.Sprintf("Text: template=%q; err=%q", t.Text, e.Text))
	}
	if t.UserMsg != "" && t.UserMsg != e.UserMsg {
		diff = append(diff, fmt.Sprintf("UserMsg: template=%q; err=%q", t.UserMsg, e.UserMsg))
	}
	if t.Data != nil && !reflect.DeepEqual(t.Data, e.Data) {
		diff = append(diff, fmt.Sprintf(`Data: template="%#v"; err="%#v"`, t.Data, e.Data))
	}
	if t.Err == nil {
		return diff
	}

	switch t.Err.(type) {
	case *Error:
		cdiff := Diff(t.Err, e.Err)
		if len(cdiff) == 0 {
			return diff
		}

		diff = append(diff, "Diff(template.Err, err.Err)=")
		for _, d := range cdiff {
			diff = append(diff, "\t"+d)
		}

	default:
		if e.Err == nil || e.Err.Error() != t.Err.Error() {
			diff = append(diff, fmt.Sprintf(`Err(nested): template="%v"(type %[1]T); err="%v"(type %[2]T)`, t.Err, e.Err))
		}
	}

	return diff
}
