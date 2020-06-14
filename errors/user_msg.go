package errors

import "errors"

// UserMsg returns the first message suitable to be shown to the
// end-user in the error chain.
func UserMsg(err error) string {
	if err == nil {
		return ""
	}

	if e, ok := err.(*Error); ok && e.UserMsg != "" {
		return e.UserMsg
	}

	return UserMsg(errors.Unwrap(err))
}
