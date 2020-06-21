package errors

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/sudo-suhas/xgo"
)

// OptionFunc is a type of constructor option for E(...)
type Option interface {
	Apply(*Error)
}

// OptionFunc type is an adapter to allow the use of ordinary functions
// as error constructor options. If f is a function with the appropriate
// signature, OptionFunc(f) is an Option that calls f.
type OptionFunc func(*Error)

// Apply calls f(e).
func (f OptionFunc) Apply(e *Error) {
	f(e)
}

// Options turns a list of Option instances into an Option.
func Options(opts ...Option) Option {
	return OptionFunc(func(e *Error) {
		for _, opt := range opts {
			opt.Apply(e)
		}
	})
}

// WithOp sets the Op on the Error instance.
func WithOp(op xgo.Op) Option {
	return OptionFunc(func(e *Error) {
		e.Op = op
	})
}

// WithUserMsg sets the UserMsg on the Error instance. msg should be
// suitable to be exposed to the end user.
func WithUserMsg(msg string) Option {
	return OptionFunc(func(e *Error) {
		e.UserMsg = msg
	})
}

// WithText sets the Text on the Error instance. It is treated as an
// error message and should not be exposed to the end user.
func WithText(text string) Option {
	return OptionFunc(func(e *Error) {
		e.Text = text
	})
}

// WithTextf sets the formatted Text on the Error instance. It is
// treated as an error message and should not be exposed to the end
// user.
func WithTextf(format string, args ...interface{}) Option {
	return WithText(fmt.Sprintf(format, args...))
}

// WithErr sets the Err on the Error instance.
func WithErr(err error) Option {
	return OptionFunc(func(e *Error) {
		if err, ok := err.(*Error); ok {
			cp := *err // Make a copy
			e.Err = &cp
			return
		}

		e.Err = err
	})
}

// WithData sets the Data on the Error instance. It can be any
// arbitrary value associated with the error.
func WithData(data interface{}) Option {
	return OptionFunc(func(e *Error) {
		e.Data = data
	})
}

// WithToJSON sets ToJSON on the Error instance. It defines the
// conversion of Error instance to a JSON value.
func WithToJSON(f JSONFunc) Option {
	return OptionFunc(func(e *Error) {
		e.ToJSON = f
	})
}

// WithResp sets the Text, Kind, Data on the Error instance.
//
// HTTP method combined with the request path and the response status is
// set as the Text. It is not recommended to set the request path as the
// Op since this URL can include an ID for some entity.
//
// The response status code is interpolated to the Kind using
// KindFromStatus.
//
// The response body is set as the Data. Special handling is included
// for detecting and preserving JSON response.
func WithResp(resp *http.Response) Option {
	return OptionFunc(func(e *Error) {
		e.Kind = KindFromStatus(resp.StatusCode)

		req := resp.Request
		e.Text = fmt.Sprintf("[%s] %s: %s", req.Method, req.URL.RequestURI(), resp.Status)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return
		}

		if isJSONContent(resp.Header.Get("Content-Type")) && isValidJSON(body) {
			e.Data = (json.RawMessage)(body)
		} else {
			e.Data = (string)(body)
		}
	})
}

// Source: https://github.com/go-resty/resty/blob/v2.2.0/client.go#L64
var jsonCheck = regexp.MustCompile(`(?i:(application|text)/(json|.*\+json|json\-.*)(;|$))`)

func isJSONContent(ct string) bool { return jsonCheck.MatchString(ct) }
func isValidJSON(data []byte) bool { return json.Unmarshal(data, &json.RawMessage{}) == nil }
