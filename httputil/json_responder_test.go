package httputil_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/sudo-suhas/xgo/errors"
	"github.com/sudo-suhas/xgo/httputil"
)

func TestJSONResponderRespond(t *testing.T) {
	cases := []struct {
		name string
		jr   httputil.JSONResponder
		v    interface{}
		want response
	}{
		{
			name: "WithValue",
			v:    Person{Name: "Donald", Age: 33},
			want: response{
				status:  http.StatusOK,
				headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
				body:    []byte(`{"Name": "Donald", "Age": 33, "V": null}`),
			},
		},
		{
			name: "WithNil",
			want: response{status: http.StatusOK},
		},
		{
			name: "WithJSONerValue",
			v:    personJSONer{Name: "Donald", Age: 33},
			want: response{
				status:  http.StatusOK,
				headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
				body:    []byte(`{"name": "Donald", "age": 33}`),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tc.jr.Respond(nil, rec, tc.v)

			matchResponse(t, rec.Result(), tc.want)
		})
	}

	t.Run("ObserveEncodeError", func(t *testing.T) {
		var callCnt int
		v := marshalFailer{err: errors.E(errors.WithOp("marshal"), errors.WithText("fail"))}
		jr := httputil.JSONResponder{
			ErrObservers: []httputil.ErrorObserverFunc{
				func(r *http.Request, err error) {
					callCnt++
					errStr := "json: error calling MarshalJSON for type httputil_test.marshalFailer: marshal: fail"
					if err == nil || err.Error() != errStr {
						t.Errorf(`ErrorObserver.err="%v"; want=%q`, err, errStr)
					}
				},
			},
		}
		rec := httptest.NewRecorder()
		jr.Respond(nil, rec, v)

		if callCnt != 1 {
			t.Errorf("ErrObserver call count=%d; want 1", callCnt)
		}
		matchResponse(t, rec.Result(), response{
			status:  http.StatusOK,
			headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
		})
	})
}

func TestJSONResponderRespondWithStatus(t *testing.T) {
	var jr httputil.JSONResponder
	rec := httptest.NewRecorder()
	jr.RespondWithStatus(nil, rec, http.StatusCreated, Person{Name: "Donald", Age: 33})
	matchResponse(t, rec.Result(), response{
		status:  http.StatusCreated,
		headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
		body:    []byte(`{"Name": "Donald", "Age": 33, "V": null}`),
	})
}

func TestJSONResponderError(t *testing.T) {
	cases := []struct {
		name string
		jr   httputil.JSONResponder
		err  error
		want response
	}{
		{
			name: "WithStatusCoderError",
			err:  errors.E(errors.InvalidInput, errors.WithUserMsg("Don't try again")),
			want: response{
				status:  http.StatusBadRequest,
				headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
				body:    []byte(`{"success":false,"msg":"Don't try again","errors":[{"code":"INVALID_INPUT","error":"invalid input","msg":"Don't try again"}]}`),
			},
		},
		{
			name: "WithMultiErrors",
			err: errors.E(
				errors.PermissionDenied,
				errors.WithUserMsg("Nice try"),
				errors.WithToJSON(func(*errors.Error) interface{} { return []string{"this", "that"} }),
			),
			want: response{
				status:  http.StatusForbidden,
				headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
				body:    []byte(`{"success":false,"msg":"Nice try","errors":["this","that"]}`),
			},
		},
		{
			name: "WithOpaqueError",
			err:  fmt.Errorf("deal with it"),
			want: response{
				status:  http.StatusInternalServerError,
				headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
				body:    []byte(`{"success":false,"msg":"","errors":null}`),
			},
		},
		{
			name: "WithErrToRespBody",
			jr:   httputil.JSONResponder{ErrToRespBody: func(err error) interface{} { return json.RawMessage(`{"no":"ok"}`) }},
			err:  errors.E(errors.PermissionDenied, errors.WithUserMsg("Nice try")),
			want: response{
				status:  http.StatusForbidden,
				headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
				body:    []byte(`{"no":"ok"}`),
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			tc.jr.Error(nil, rec, tc.err)

			matchResponse(t, rec.Result(), tc.want)
		})
	}
	t.Run("ErrObservers", func(t *testing.T) {
		var callCnt int
		respondErr := errors.E(
			errors.WithOp("marshal"), errors.InvalidInput, errors.WithText("fail"),
		)
		jr := httputil.JSONResponder{
			ErrObservers: []httputil.ErrorObserverFunc{
				func(r *http.Request, err error) {
					callCnt++
					if !errors.Match(respondErr, err) {
						t.Errorf("ErrorObserver.err diff= %s", errorDiff(respondErr, err))
					}
				},
				func(*http.Request, error) { callCnt++ },
			},
		}
		rec := httptest.NewRecorder()
		jr.Error(nil, rec, respondErr)

		if callCnt != 2 {
			t.Errorf("ErrObserver call count=%d; want 2", callCnt)
		}
		matchResponse(t, rec.Result(), response{
			status:  http.StatusBadRequest,
			headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
			body:    []byte(`{"success":false,"msg":"","errors":[{"code":"INVALID_INPUT","error":"invalid input","msg":""}]}`),
		})
	})
}

func TestJSONResponderErrorWithStatus(t *testing.T) {
	var jr httputil.JSONResponder
	rec := httptest.NewRecorder()
	err := errors.E(errors.InvalidInput, errors.WithUserMsg("Don't try again"))
	jr.ErrorWithStatus(nil, rec, http.StatusConflict, err)
	matchResponse(t, rec.Result(), response{
		status:  http.StatusConflict,
		headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
		body:    []byte(`{"success":false,"msg":"Don't try again","errors":[{"code":"INVALID_INPUT","error":"invalid input","msg":"Don't try again"}]}`),
	})
}

func matchResponse(t *testing.T, got *http.Response, want response) {
	t.Helper()
	if got.StatusCode != want.status {
		t.Errorf("StatusCode=%d; want=%d", got.StatusCode, want.status)
	}

	if !reflect.DeepEqual(headers(got), want.headers) {
		t.Errorf("Headers=%q; want=%q", headers(got), want.headers)
	}

	// (*httptest.ResponseRecorder).Result()
	// The Response.Body is guaranteed to be non-nil and Body.Read call is
	// guaranteed to not return any error other than io.EOF.
	body, _ := ioutil.ReadAll(got.Body)
	body = bytes.TrimSpace(body)
	if len(want.body) > 0 {
		ok, err := jsonBytesEqual(body, want.body)
		if err != nil {
			t.Fatalf("jsonBytesEqual()=%q", err)
		}

		if !ok {
			t.Errorf("Body=%s; want=%s", body, want.body)
		}
	} else if len(body) > 0 {
		t.Errorf("Body=%s; want=<empty>", body)
	}
}

type response struct {
	status  int
	headers map[string]string
	body    []byte
}

// jsonBytesEqual compares the JSON in two byte slices.
func jsonBytesEqual(a, b []byte) (bool, error) {
	// See https://stackoverflow.com/questions/32408890/how-to-compare-two-json-requests
	var j, j2 interface{}
	if err := json.Unmarshal(a, &j); err != nil {
		return false, err
	}
	if err := json.Unmarshal(b, &j2); err != nil {
		return false, err
	}
	return reflect.DeepEqual(j2, j), nil
}

func headers(r *http.Response) map[string]string {
	if len(r.Header) == 0 {
		return nil
	}

	h := make(map[string]string)
	for k := range r.Header {
		h[k] = r.Header.Get(k)
	}
	return h
}

type personJSONer Person

func (p personJSONer) JSON() interface{} {
	return map[string]interface{}{"name": p.Name, "age": p.Age}
}

type marshalFailer struct{ err error }

func (m marshalFailer) MarshalJSON() ([]byte, error) { return nil, m.err }
