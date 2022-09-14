package errors

// InternalDetails is the internal details populated from the error
// instance.
type InternalDetails struct {
	Ops   []string    `json:"ops,omitempty"`
	Kind  Kind        `json:"kind,omitempty"`
	Error string      `json:"error"`
	Data  interface{} `json:"data,omitempty"`
}

// Details constructs and yields the details of the error by traversing
// the error chain.
func (e *Error) Details() InternalDetails {
	var dd []interface{}
	walk(e, func(err *Error) {
		if err.Data != nil {
			dd = append(dd, err.Data)
		}
	})

	var data interface{}
	if len(dd) == 1 {
		data = dd[0]
	}
	if len(dd) > 1 {
		data = dd
	}

	return InternalDetails{
		Ops:   e.Ops(),
		Kind:  WhatKind(e),
		Error: e.Error(),
		Data:  data,
	}
}
