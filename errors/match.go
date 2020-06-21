package errors

import "reflect"

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
	e1, ok := err1.(*Error)
	if !ok {
		return false
	}
	e2, ok := err2.(*Error)
	if !ok {
		return false
	}
	if e1.Op != "" && e1.Op != e2.Op {
		return false
	}
	if e1.Kind != Unknown && e1.Kind != e2.Kind {
		return false
	}
	if e1.Text != "" && e1.Text != e2.Text {
		return false
	}
	if e1.UserMsg != "" && e1.UserMsg != e2.UserMsg {
		return false
	}
	if e1.Data != nil && !reflect.DeepEqual(e1.Data, e2.Data) {
		return false
	}
	if e1.Err != nil {
		if _, ok := e1.Err.(*Error); ok {
			return Match(e1.Err, e2.Err)
		}
		if e2.Err == nil || e2.Err.Error() != e1.Err.Error() {
			return false
		}
	}

	return true
}
