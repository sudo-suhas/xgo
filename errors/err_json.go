package errors

// JSONFunc converts the Error instance to a JSON value. It is
// recommended to return one of the following:
//
//   - map[string]interface{}
//   - []map[string]inteface{}
//   - CustomType
//
// In certain cases, such as validation errors, it is reasonable to
// expand a single error instance into multiple 'Objects'.
//
// Care must be taken not to expose details internal to the application.
// This is for two reasons, namely security and user experience.
type JSONFunc func(*Error) interface{}

// JSON is the default implementation of representing the error as a
// JSON value.
func (e *Error) JSON() interface{} {
	if e.ToJSON != nil {
		return e.ToJSON(e)
	}

	k := WhatKind(e)
	return map[string]interface{}{
		"code":  k.Code,
		"error": k.String(),
		"msg":   UserMsg(e),
	}
}
