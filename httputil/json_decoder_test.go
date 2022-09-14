package httputil_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"github.com/sudo-suhas/xgo/errors"
	"github.com/sudo-suhas/xgo/httputil"
)

func TestJSONDecoderDecode(t *testing.T) {
	var (
		method = http.MethodGet
		url    = "http://host.com/route"
	)
	cases := []struct {
		name    string
		j       httputil.JSONDecoder
		r       request
		v       interface{}
		want    interface{}
		wantErr error
	}{
		{
			name: "Success",
			r: request{
				method:  method,
				url:     url,
				headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
				body:    `{ "name": "Donald", "age": 33, "v": 10 }`,
			},
			v:    &Person{},
			want: &Person{Name: "Donald", Age: 33, V: float64(10)},
		},
		{
			name: "SuccessWithSkipCheckContentType",
			j:    httputil.JSONDecoder{SkipCheckContentType: true},
			r: request{
				method: method,
				url:    url,
				body:   `{ "name": "Donald", "age": 33 }`,
			},
			v:    &Person{},
			want: &Person{Name: "Donald", Age: 33},
		},
		{
			name: "SuccessWithUseNumber",
			j:    httputil.JSONDecoder{UseNumber: true},
			r: request{
				method:  method,
				url:     url,
				headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
				body:    `{ "name": "Donald", "age": 33, "v": 10 }`,
			},
			v:    &Person{},
			want: &Person{Name: "Donald", Age: 33, V: json.Number("10")},
		},
		{
			name: "ContentTypeNotPresent",
			r: request{
				method: method,
				url:    url,
				body:   `{ "name": "Donald", "age": 33 }`,
			},
			v: &Person{},
			wantErr: errors.E(
				errors.WithOp("JSONDecoder.Decode"),
				httputil.ErrKindUnsupportedMediaType,
				errors.WithText("Content-Type header '' is not application/json"),
			),
		},
		{
			name: "ContentTypeNotAccepted",
			r: request{
				method:  method,
				url:     url,
				headers: map[string]string{"Content-Type": "text/html; charset=utf-8"},
				body:    `{ "name": "Donald", "age": 33 }`,
			},
			v: &Person{},
			wantErr: errors.E(
				errors.WithOp("JSONDecoder.Decode"),
				httputil.ErrKindUnsupportedMediaType,
				errors.WithText("Content-Type header 'text/html; charset=utf-8' is not application/json"),
			),
		},
		{
			name: "SyntaxError",
			r: request{
				method:  method,
				url:     url,
				headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
				body:    `{ "name": "Donald", "age": }`,
			},
			v: &Person{},
			wantErr: errors.E(
				errors.WithOp("JSONDecoder.Decode"),
				errors.InvalidInput,
				errors.WithUserMsg("Request body contains badly-formed JSON (at position 28)"),
			),
		},
		{
			name: "UnexpectedEOF",
			r: request{
				method:  method,
				url:     url,
				headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
				body:    `{ "name": " }`,
			},
			v: &Person{},
			wantErr: errors.E(
				errors.WithOp("JSONDecoder.Decode"),
				errors.InvalidInput,
				errors.WithUserMsg("Request body contains badly-formed JSON"),
			),
		},
		{
			name: "TypeError",
			r: request{
				method:  method,
				url:     url,
				headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
				body:    `{ "name": "Donald", "age": "middle" }`,
			},
			v: &Person{},
			wantErr: errors.E(
				errors.WithOp("JSONDecoder.Decode"),
				errors.InvalidInput,
				errors.WithUserMsg("Request body contains an invalid value for the 'Age' field (at position 35)"),
			),
		},
		{
			name: "UnknownField",
			j:    httputil.JSONDecoder{DisallowUnknownFields: true},
			r: request{
				method:  method,
				url:     url,
				headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
				body:    `{ "name": "Donald", "age": 10, "height": 175 }`,
			},
			v: &Person{},
			wantErr: errors.E(
				errors.WithOp("JSONDecoder.Decode"),
				errors.InvalidInput,
				errors.WithUserMsg("Request body contains unknown field 'height'"),
			),
		},
		{
			name: "EmptyBody",
			r: request{
				method:  method,
				url:     url,
				headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
				body:    ``,
			},
			v: &Person{},
			wantErr: errors.E(
				errors.WithOp("JSONDecoder.Decode"),
				errors.InvalidInput,
				errors.WithUserMsg("Request body must not be empty"),
			),
		},
		{
			name: "MultipleJSONObjects",
			r: request{
				method:  method,
				url:     url,
				headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
				body:    `{ "name": "Donald", "age": 33 }` + "\n" + `{ "name": "Puddy", "age": 33 }`,
			},
			v: &Person{},
			wantErr: errors.E(
				errors.WithOp("JSONDecoder.Decode"),
				errors.InvalidInput,
				errors.WithUserMsg("Request body must only contain a single JSON object"),
			),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			r, err := tc.r.build()
			if err != nil {
				t.Fatalf("http.NewRequest: %s", err)
			}

			err = tc.j.Decode(r, tc.v)
			if !matchErrors(tc.wantErr, err) {
				t.Errorf(
					"JSONDecoder.Decode() error diff: %s", errorDiff(tc.wantErr, err),
				)
				return
			}

			if err == nil && !reflect.DeepEqual(tc.want, tc.v) {
				t.Errorf("\nJSONDecoder.Decode()=%#v \nwant %#v", tc.v, tc.want)
			}
		})
	}

	t.Run("RequestBodyTooLarge", func(t *testing.T) {
		r, err := request{
			method:  method,
			url:     url,
			headers: map[string]string{"Content-Type": "application/json; charset=utf-8"},
			body:    `{ "name": "Donald", "age": 33 }`,
		}.build()
		if err != nil {
			t.Fatalf("http.NewRequest: %s", err)
		}

		r.Body = http.MaxBytesReader(httptest.NewRecorder(), r.Body, 1)
		err = httputil.JSONDecoder{}.Decode(r, &Person{})
		want := errors.E(errors.WithOp("JSONDecoder.Decode"), httputil.ErrKindRequestEntityTooLarge)
		if !errors.Match(want, err) {
			t.Errorf("JSONDecoder.Decode() error diff: %s", errorDiff(want, err))
		}
	})
}

func matchErrors(want, got error) bool {
	if want == nil {
		return got == nil
	}
	return errors.Match(want, got)
}

type request struct {
	method  string
	url     string
	headers map[string]string
	body    string
}

func (req request) build() (*http.Request, error) {
	r, err := http.NewRequest(req.method, req.url, strings.NewReader(req.body))
	if err != nil {
		return nil, err
	}

	for key, value := range req.headers {
		r.Header.Set(key, value)
	}

	return r, nil
}

type Person struct {
	Name string
	Age  int
	V    interface{}
}

func errorDiff(template, err error) string {
	return "\n- " + strings.Join(errors.Diff(template, err), "\n- ")
}
