package errors

import (
	"bytes"

	"github.com/sudo-suhas/xgo"
)

// Package types are inspired by the following articles:
// - https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html
// - https://middlemost.com/failure-is-your-domain/
// - https://about.sourcegraph.com/go/gophercon-2019-handling-go-errors

// Error is the type that implements the error interface. An Error value
// may leave some values unset.
//
// If the error is printed, only those items that have been set to
// non-zero values will appear in the result.
//
// If Kind is not specified or Unknown, we try to set it to the Kind of
// the underlying error.
type Error struct {
	// Op is the operation being performed, usually the name of the method
	// being invoked (Get, Put, etc.).
	Op xgo.Op

	// Kind is the class of error, such as permission failure,
	// or "Unknown" if its class is unknown or irrelevant.
	Kind Kind

	// Text is the error string. Text is not expected to be suitable to
	// be shown to the end user.
	Text string

	// UserMsg is the error message suitable to be shown to the end
	// user.
	UserMsg string

	// Data is arbitrary value associated with the error. Data is not
	// expected to be suitable to be shown to the end user.
	Data interface{}

	// Err is the underlying error that triggered this one, if any.
	Err error

	// ToJSON is used to override the default implementation of
	// converting the Error instance into a JSON value. Optional.
	ToJSON JSONFunc
}

// E builds an error value with the provided options.
//
// 	if err := svc.ProcessSomething(); err != nil {
// 		return errors.E(errors.WithOp(op), errors.WithErr(err))
// 	}
func E(opt Option, opts ...Option) error {
	var e Error
	for _, opt := range prepend(opt, opts) {
		opt.Apply(&e)
	}

	e.promoteFields()
	return &e
}

func (e *Error) Error() string {
	b := new(bytes.Buffer)

	writePaddedStr(b, string(e.Op))
	if e.Kind != Unknown {
		writePaddedStr(b, e.Kind.String())
	}
	writePaddedStr(b, e.Text)
	if e.Err != nil {
		writePaddedStr(b, e.Err.Error())
	}

	if b.Len() == 0 {
		return "no error"
	}
	return b.String()
}

// Ops returns the "stack" of operations for the error.
func (e *Error) Ops() []xgo.Op {
	var ops []xgo.Op
	walk(e, func(err *Error) {
		if err.Op != "" {
			ops = append(ops, err.Op)
		}
	})
	return ops
}

// GetKind implements the GetKind interface.
func (e *Error) GetKind() Kind { return e.Kind }

// Unwrap unpacks wrapped errors. It is used by functions errors.Is and
// errors.As.
func (e *Error) Unwrap() error { return e.Err }

func (e *Error) promoteFields() {
	prev, ok := e.Err.(*Error)
	if !ok {
		return
	}

	// The previous error was also one of ours. Suppress duplications
	// so the message won't contain the same op, kind etc.
	if e.Op == prev.Op {
		prev.Op = ""
	}
	if prev.Kind == e.Kind {
		prev.Kind = Unknown
	}
	if prev.UserMsg == e.UserMsg {
		prev.UserMsg = ""
	}
	if prev.Text == e.Text {
		prev.Text = ""
	}

	// If this error has Op/UserMsg/Kind/Data/ToJSON unset, pull up the
	// inner one.
	if e.Op == "" {
		e.Op, prev.Op = prev.Op, ""
	}
	if e.Kind == Unknown {
		e.Kind, prev.Kind = prev.Kind, Unknown
	}
	if e.UserMsg == "" {
		e.UserMsg, prev.UserMsg = prev.UserMsg, ""
	}
	if e.Data == nil {
		e.Data, prev.Data = prev.Data, nil
	}
	if e.ToJSON == nil {
		e.ToJSON, prev.ToJSON = prev.ToJSON, nil
	}

	if prev.Op != "" || prev.Kind != Unknown {
		// If Op/Kind is present, neither Text nor Err can be promoted up.
		return
	}
	if e.Text == "" {
		e.Text, prev.Text = prev.Text, ""
	}

	if prev.isZero() {
		e.Err = nil
	} else if hasOnlyErr(*prev) {
		e.Err = prev.Err
	}
}

func (e *Error) isZero() bool {
	return e.Op == "" &&
		e.Kind == Unknown &&
		e.Text == "" &&
		e.Err == nil &&
		e.UserMsg == "" &&
		e.Data == nil &&
		e.ToJSON == nil
}

func walk(e *Error, f func(*Error)) {
	for e != nil {
		f(e)

		var ok bool
		if e, ok = e.Err.(*Error); !ok {
			return
		}
	}
}

func prepend(head Option, body []Option) []Option {
	opts := make([]Option, len(body)+1)
	opts[0] = head
	copy(opts[1:], body)
	return opts
}

func writePaddedStr(b *bytes.Buffer, s string) {
	if s == "" {
		return
	}
	pad(b, ": ")
	b.WriteString(s)
}

// pad appends str to the buffer if the buffer already has some data.
func pad(b *bytes.Buffer, str string) {
	if b.Len() == 0 {
		return
	}
	b.WriteString(str)
}

func hasOnlyErr(e Error) bool {
	e.Err = nil
	return e.isZero()
}
