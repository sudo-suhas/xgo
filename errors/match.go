package errors

import (
	"fmt"
	"reflect"
)

// Match compares its two error arguments. It can be used to check for
// expected errors in tests. Both arguments must have underlying type
// *Error or Match will return false. Otherwise it returns true iff
// every non-zero element of the first error is equal to the
// corresponding element of the second. If the Err field is a *Error,
// Match recurs on that field; otherwise it compares the strings
// returned by the Error methods. Elements that are in the second
// argument but not present in the first are ignored.
//
// For example,
//	errors.Match(errors.E(errors.WithOp("service.MakeBooking"), errors.PermissionDenied), err)
// tests whether err is an Error with Kind=PermissionDenied and
// Op=service.MakeBooking.
func Match(err1, err2 error) bool {
	return len(Diff(err1, err2)) == 0
}

// Diff compares and returns the diff between two error arguments. It
// can be used to extract the exact details of difference between
// expected and got errors in tests. It returns an empty arrat iff
// both arguments are of type *Error and every non-zero element of the
// first error is equal to the corresponding element of the second. If
// the Err field is a *Error, Diff recurs on that field; otherwise it
// compares the strings returned by the Error methods. Elements that are
// in the second argument but not present in the first are ignored.
//
// For example,
// 	errors.Diff(errors.E(errors.WithOp("service.MakeBooking"), errors.PermissionDenied), err)
// tests whether err is an Error with Kind=PermissionDenied and
// Op=service.MakeBooking.
func Diff(err1, err2 error) []string { //nolint: gocognit
	e1, ok1 := err1.(*Error)
	e2, ok2 := err2.(*Error)
	if !ok1 || !ok2 {
		return []string{fmt.Sprintf(`template="%v"(type %[1]T); err="%v"(type %[2]T)`, err1, err2)}
	}

	var a []string
	if e1.Op != "" && e1.Op != e2.Op {
		a = append(a, fmt.Sprintf("Op: template=%q; err=%q", e1.Op, e2.Op))
	}
	if e1.Kind != Unknown && e1.Kind != e2.Kind {
		a = append(a, fmt.Sprintf("Kind: template=%q; err=%q", e1.Kind, e2.Kind))
	}
	if e1.Text != "" && e1.Text != e2.Text {
		a = append(a, fmt.Sprintf("Text: template=%q; err=%q", e1.Text, e2.Text))
	}
	if e1.UserMsg != "" && e1.UserMsg != e2.UserMsg {
		a = append(a, fmt.Sprintf("UserMsg: template=%q; err=%q", e1.UserMsg, e2.UserMsg))
	}
	if e1.Data != nil && !reflect.DeepEqual(e1.Data, e2.Data) {
		a = append(a, fmt.Sprintf(`Data: template="%#v"; err="%#v"`, e1.Data, e2.Data))
	}
	if e1.Err == nil {
		return a
	}

	switch e1.Err.(type) {
	case *Error:
		b := Diff(e1.Err, e2.Err)
		if len(b) == 0 {
			return a
		}

		a = append(a, "Diff(template.Err, err.Err)=")
		for _, d := range b {
			a = append(a, "\t"+d)
		}

	default:
		if e2.Err == nil || e2.Err.Error() != e1.Err.Error() {
			a = append(a, fmt.Sprintf(`Err(nested): template="%v"(type %[1]T); err="%v"(type %[2]T)`, e1.Err, e2.Err))
		}
	}

	return a
}
