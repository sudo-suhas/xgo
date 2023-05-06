package errors

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"runtime"
	"strconv"
	"testing"
)

func TestOptions(t *testing.T) {
	got := E(Options(WithOp("Get"), InvalidInput, WithText("stoinks")))
	want := E(WithOp("Get"), InvalidInput, WithText("stoinks"))
	if !Match(want, got) {
		t.Errorf("E(...)=%q; want %q", got, want)
	}
}

func TestOption(t *testing.T) {
	cases := []struct {
		name string
		opt  Option
		want error
	}{
		{"Op", WithOp("Get"), &Error{Op: "Get"}},
		{"UserMsg", WithUserMsg("Deal with it!"), &Error{UserMsg: "Deal with it!"}},
		{"Text", WithText("stoinks"), &Error{Text: "stoinks"}},
		{"Textf", WithTextf("stoinks: %s: %d", "zoinks", 420), &Error{Text: "stoinks: zoinks: 420"}},
		{"Err", WithErr(E(WithOp("Get"))), &Error{Op: "Get"}},
		{"ErrNonError", WithErr(errors.New("invalid input")), &Error{Err: errors.New("invalid input")}},
		{"Data", WithData(420), &Error{Data: 420}},
		{
			"Fields",
			Fields{Op: "Get", Kind: InvalidInput, Text: "stoinks", UserMsg: "Deal with it!", Data: 420, Err: E(WithOp("root"))},
			&Error{Op: "Get", Kind: InvalidInput, Text: "stoinks", UserMsg: "Deal with it!", Data: 420, Err: &Error{Op: "root"}},
		},
		{
			"FieldsZero", // Only no-zero value fields are used.
			Options(WithOp("Get"), InvalidInput, WithText("stoinks"), WithUserMsg("Deal with it!"), WithData(420), WithErr(E(WithOp("root"))), Fields{}),
			&Error{Op: "Get", Kind: InvalidInput, Text: "stoinks", UserMsg: "Deal with it!", Data: 420, Err: &Error{Op: "root"}},
		},
	}
	for _, tc := range cases {
		t.Run("With"+tc.name, func(t *testing.T) {
			if got := E(tc.opt); !Match(tc.want, got) {
				t.Errorf("E(...)=%q; want %q", got, tc.want)
			}
		})
	}
}

func TestWithToJSON(t *testing.T) {
	e := E(WithToJSON(nilJSON)).(*Error)
	if e.ToJSON == nil {
		t.Errorf("Error.ToJSON=nil; want %s", funcName(nilJSON))
		return
	}
	if funcName(nilJSON) != funcName(e.ToJSON) {
		t.Errorf("Error.ToJSON=%s; want %s", funcName(e.ToJSON), funcName(nilJSON))
	}
}

//nolint:misspell
func TestWithResp(t *testing.T) {
	cases := []struct {
		name string
		res  *http.Response
		want error
	}{
		{
			name: "HTMLRes",
			res: newResponse(
				httptest.NewRequest(http.MethodGet, "https://developer.mozilla.org/en-US/404", nil),
				http.StatusNotFound,
				"text/html; charset=utf-8",
				html404,
			),
			want: E(NotFound, WithText("[GET] /en-US/404: 404 Not Found"), WithData(html404)),
		},
		{
			name: "JSONRes",
			res: newResponse(
				httptest.NewRequest(http.MethodPut, "https://api.github.com/user/starred/sudo-suhas/xgo", nil),
				http.StatusUnauthorized,
				"application/json; charset=utf-8",
				json401,
			),
			want: E(Unauthenticated, WithText("[PUT] /user/starred/sudo-suhas/xgo: 401 Unauthorized"), WithData(json.RawMessage(json401))),
		},
		{
			name: "InvalidJSONRes",
			res: newResponse(
				httptest.NewRequest(http.MethodGet, "https://api.terrible.com/orders", nil),
				http.StatusUnauthorized,
				"application/json; charset=utf-8",
				"User not logged in",
			),
			want: E(Unauthenticated, WithText("[GET] /orders: 401 Unauthorized"), WithData("User not logged in")),
		},
		{
			name: "BadContentType",
			res: newResponse(
				httptest.NewRequest(http.MethodGet, "https://api.terrible.com/orders", nil),
				http.StatusUnauthorized,
				"application/jsn; charset=utf-8",
				`{"message": "Requires authentication"}`,
			),
			want: E(Unauthenticated, WithText("[GET] /orders: 401 Unauthorized"), WithData(`{"message": "Requires authentication"}`)),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := E(WithResp(tc.res)); !Match(tc.want, got) {
				t.Errorf("E(...)=%q; want %q", got, tc.want)
			}
		})
	}
}

func nilJSON(*Error) interface{} { return nil }

func funcName(f interface{}) string {
	// https://stackoverflow.com/questions/7052693/how-to-get-the-name-of-a-function-in-go
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

func newResponse(req *http.Request, status int, contentType, body string) *http.Response {
	header := make(http.Header, 1)
	header.Set("Content-Type", contentType)
	return &http.Response{
		Status:        strconv.Itoa(status) + " " + http.StatusText(status),
		StatusCode:    status,
		Proto:         "HTTP/1.1",
		ProtoMajor:    1,
		ProtoMinor:    1,
		Body:          io.NopCloser(bytes.NewBufferString(body)),
		ContentLength: int64(len(body)),
		Request:       req,
		Header:        header,
	}
}

const (
	html404 = `<!doctype html>
<html lang="en">

<head>
  <meta charset="utf-8">
  <title>Page Not Found</title>
</head>

<body>
  <h1>Page Not Found</h1>
  <p>Sorry, but the page you were trying to view does not exist.</p>
</body>

</html>`

	json401 = `{
  "message": "Requires authentication",
  "documentation_url": "https://developer.github.com/v3/#authentication"
}`
)
