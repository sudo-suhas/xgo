package httputil

import (
	"encoding/json"
	"net/http"
	"reflect"

	"github.com/sudo-suhas/xgo"
	"github.com/sudo-suhas/xgo/errors"
)

// ErrorObserverFunc takes some action when an error occurs during
// request processing.
//
// 	func errLogger(r *http.Request, err error) {
// 		var e *errors.Error
// 		if !errors.As(err, &e) {
// 			httplog.LogEntrySetField(r, "error", err.Error())
// 			return
// 		}
//
// 		httplog.LogEntrySetField(r, "error_details", e.Details())
// 	}
//
type ErrorObserverFunc func(r *http.Request, err error)

// JSONResponder responds with the value or error encoded as JSON.
type JSONResponder struct {
	// ErrToRespBody converts the error to the response body. Optional.
	ErrToRespBody func(error) interface{}

	// ErrObservers are notified of errors for responses sent via
	// JSONResponder.Error and JSONResponder.ErrorWithStatus.
	ErrObservers []ErrorObserverFunc
}

// Respond encodes v as JSON and writes the response with status
// '200: OK'. Only the HTTP status is written as response if v is nil.
// Furthermore, interface upgrade to xgo.JSON is supported for v.
func (jr *JSONResponder) Respond(r *http.Request, w http.ResponseWriter, v interface{}) {
	jr.RespondWithStatus(r, w, http.StatusOK, v)
}

// RespondWithStatus encodes the value as JSON and writes the response
// with the specified status code. Only HTTP status is written as the
// response if v is nil. Furthermore, interface upgrade to xgo.JSON is
// supported for v.
func (jr *JSONResponder) RespondWithStatus(r *http.Request, w http.ResponseWriter, status int, v interface{}) {
	if v == nil {
		w.WriteHeader(status)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)

	body := v
	if j, ok := v.(xgo.JSONer); ok {
		body = j.JSON()
	}

	if err := json.NewEncoder(w).Encode(body); err != nil {
		jr.observeError(r, err)
	}
}

// Error writes the error response. The status code and response body
// are constructed from the error. ErrToResponseBody can be used to
// define/override the response body structure.
func (jr *JSONResponder) Error(r *http.Request, w http.ResponseWriter, err error) {
	jr.ErrorWithStatus(r, w, errors.StatusCode(err), err)
}

// ErrorWithStatus writes the error response. The response body is
// constructed from the error. ErrToResponseBody can be used to
// define/override the response body structure.
func (jr *JSONResponder) ErrorWithStatus(r *http.Request, w http.ResponseWriter, status int, err error) {
	jr.observeError(r, err)

	jr.RespondWithStatus(r, w, status, jr.convertErrorToBody(err))
}

func (jr *JSONResponder) observeError(r *http.Request, err error) {
	for _, f := range jr.ErrObservers {
		f(r, err)
	}
}

func (jr *JSONResponder) convertErrorToBody(err error) interface{} {
	if jr.ErrToRespBody != nil {
		return jr.ErrToRespBody(err)
	}

	var body struct {
		Success bool        `json:"success"`
		Msg     string      `json:"msg"`
		Errors  interface{} `json:"errors"`
	}

	body.Msg = errors.UserMsg(err)

	var j xgo.JSONer
	if errors.As(err, &j) {
		switch v := j.JSON(); reflect.TypeOf(v).Kind() {
		case reflect.Slice, reflect.Array:
			body.Errors = v

		default:
			body.Errors = []interface{}{v}
		}
	}

	return body
}
