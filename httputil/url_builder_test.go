package httputil_test

import (
	"fmt"
	"net/url"
	"reflect"
	"testing"

	"github.com/sudo-suhas/xgo/errors"
	"github.com/sudo-suhas/xgo/httputil"
)

func TestNewURLBuilderSource(t *testing.T) {
	cases := []struct {
		name    string
		baseURL string
		want    *url.URL
		wantErr error
	}{
		{
			name:    "Simple",
			baseURL: "https://api.example.com",
			want: &url.URL{
				Scheme: "https",
				Host:   "api.example.com",
			},
		},
		{
			name:    "WithQueryParams",
			baseURL: "http://example.com?limit=10&offset=120",
			want: &url.URL{
				Scheme:   "http",
				Host:     "example.com",
				RawQuery: "limit=10&offset=120",
			},
		},
		{
			name:    "WithoutScheme",
			baseURL: "api.example.com",
			want: &url.URL{
				Scheme: "http",
				Host:   "api.example.com",
			},
		},
		{
			name:    "InvalidBaseURL",
			baseURL: ":foo",
			wantErr: errors.E(
				errors.WithOp("httputil.NewURLBuilderSource"),
				errors.InvalidInput,
				errors.WithErr(fmt.Errorf(`parse ":foo": missing protocol scheme`)),
			),
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			b, err := httputil.NewURLBuilderSource(tc.baseURL)
			if !matchErrors(tc.wantErr, err) {
				t.Errorf("NewURLBuilderSource() error diff: %s", errorDiff(tc.wantErr, err))
				return
			}

			if err == nil && !reflect.DeepEqual(tc.want, b.NewURLBuilder().URL()) {
				t.Errorf("NewURLBuilder().URL()=%#v \nwant %#v", b.NewURLBuilder().URL(), tc.want)
			}
		})
	}
}

func TestURLBuilder(t *testing.T) {
	b, err := httputil.NewURLBuilderSource("https://api.example.com/v1/?limit=10&mode=light")
	if err != nil {
		t.Fatalf("NewURLBuilderSource(): %s", err)
	}

	cases := []struct {
		name string
		url  *url.URL
		want *url.URL
	}{
		{
			name: "Simple",
			url:  b.NewURLBuilder().Path("users").URL(),
			want: &url.URL{
				Scheme:   "https",
				Host:     "api.example.com",
				Path:     "/v1/users",
				RawQuery: "limit=10&mode=light",
			},
		},
		{
			name: "PathParams",
			url: b.NewURLBuilder().
				Path("/users/{userID}/posts/{postID}/comments").
				PathParam("userID", "foo").
				PathParamInt("postID", 42).
				URL(),
			want: &url.URL{
				Scheme:   "https",
				Host:     "api.example.com",
				Path:     "/v1/users/foo/posts/42/comments",
				RawQuery: "limit=10&mode=light",
			},
		},
		{
			name: "DoublePathParam",
			url: b.NewURLBuilder().
				Path("/path/{p1}/{p2}/{p1}").
				PathParam("p1", "a").
				PathParam("p2", "b").
				URL(),
			want: &url.URL{
				Scheme:   "https",
				Host:     "api.example.com",
				Path:     "/v1/path/a/b/a",
				RawQuery: "limit=10&mode=light",
			},
		},
		{
			name: "EscapePathParam",
			url: b.NewURLBuilder().
				Path("/posts/{title}").
				PathParam("title", `Letters & "Special" Characters`).
				URL(),
			want: &url.URL{
				Scheme:   "https",
				Host:     "api.example.com",
				Path:     "/v1/posts/Letters%20&%20%22Special%22%20Characters",
				RawQuery: "limit=10&mode=light",
			},
		},
		{
			name: "PathParamsMap",
			url: b.NewURLBuilder().
				Path("{app}/users/{userID}/posts/{postID}/comments").
				PathParams(map[string]string{
					"app":    "myapp",
					"userID": "1",
					"postID": "42",
				}).
				URL(),
			want: &url.URL{
				Scheme:   "https",
				Host:     "api.example.com",
				Path:     "/v1/myapp/users/1/posts/42/comments",
				RawQuery: "limit=10&mode=light",
			},
		},
		{
			name: "QueryParams",
			url: b.NewURLBuilder().
				Path("/search").
				QueryParam("author_id", "foo", "bar").
				QueryParamInt("limit", 20).
				QueryParamInt("ints", 1, 3, 5, 7).
				QueryParamBool("recent", true).
				QueryParamFloat("min_rating", 4.5).
				QueryParamFloat("floats", 0, -2, 4.6735593624473).
				URL(),
			want: &url.URL{
				Scheme:   "https",
				Host:     "api.example.com",
				Path:     "/v1/search",
				RawQuery: "author_id=foo&author_id=bar&floats=0&floats=-2&floats=4.6735593624473&ints=1&ints=3&ints=5&ints=7&limit=20&min_rating=4.5&mode=light&recent=true",
			},
		},
		{
			name: "EscapeQueryParam",
			url: b.NewURLBuilder().
				Path("/search").
				QueryParam("text", "foo bar/®").
				URL(),
			want: &url.URL{
				Scheme:   "https",
				Host:     "api.example.com",
				Path:     "/v1/search",
				RawQuery: "limit=10&mode=light&text=foo+bar%2F%C2%AE",
			},
		},
		{
			name: "QueryParamValues",
			url: b.NewURLBuilder().
				QueryParams(url.Values{
					"mode":   {"dark"},
					"offset": {"20"},
				}).
				URL(),
			want: &url.URL{
				Scheme:   "https",
				Host:     "api.example.com",
				Path:     "/v1",
				RawQuery: "limit=10&mode=dark&offset=20",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !reflect.DeepEqual(tc.url, tc.want) {
				t.Errorf("URLBuilder().URL()=%#v \nwant %#v", tc.url, tc.want)
			}
		})
	}
}
