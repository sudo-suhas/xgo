# httputil [![PkgGoDev][pkg-go-dev-xgo-badge]][pkg-go-dev-xgo-httputil]

HTTP utility functions focused around decoding requests and encoding responses
in JSON.

## Table of contents

- [Usage](#usage)
  - [Decoding requests](#decoding-requests)
    - [`JSONDecoder`](#jsondecoder)
    - [Validation](#validation)
    - [Decoding query parameters](#decoding-query-parameters)
  - [Encoding responses](#encoding-responses)
    - [Encoding errors](#encoding-errors)
    - [Observing errors](#observing-errors)
  - [Building URLs](#building-urls)

## Usage

```go
import "github.com/sudo-suhas/xgo/httputil"
```

### Decoding requests

The `httputil` package defines the [`Decoder`][decoder] interface to serve as
the central building block:

```go
type Decoder interface {
	// Decode decodes the HTTP request into the given value.
	Decode(r *http.Request, v interface{}) error
}
```

#### `JSONDecoder`

[`JSONDecoder`][jsondecoder] implements this interface and can be used to parse
the request body if the content type is JSON.

```go
func CreateUserHandler(svc myapp.UserService) http.Handler {
	var (
		jsonDec   httputil.JSONDecoder
		responder httputil.JSONResponder
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var u myapp.User
		if err := jsonDec.Decode(r, &u); err != nil {
			responder.Error(r, w, err)
			return
		}

		id, err := svc.Create(r.Context(), u)
		if err != nil {
			responder.Error(r, w, err)
			return
		}

		responder.Respond(r, w, myapp.Response{
			Success: true,
			Data:    id,
		})
	})
}
```

By default, the [`JSONDecoder`][jsondecoder] looks for the `Content-Type` header
and returns an error if it is not JSON:

```
$ curl -X POST -s localhost:3000/user | jq
{
  "success": false,
  "msg": "",
  "errors": [
    {
      "code": "UNSUPPORTED_MEDIA_TYPE",
      "error": "unsupported media type",
      "msg": ""
    }
  ]
}
```

This can be disabled by setting `SkipCheckContentType` to `true` on the
[`JSONDecoder`][jsondecoder] instance.

#### Validation

Validation of input can be plugged into the decoding step using
[`ValidatingDecoderMiddleware`][validatingdecodermiddleware] with an
implementation of [`xgo.Validator`][validator]:

```go
var vd xgo.Validator = MyValidator{}
var dec httputil.Decoder
{
	dec = httputil.JSONDecoder{}
	dec = httputil.ValidatingDecoderMiddleware(vd)(dec)
}
```

#### Decoding query parameters

Decoding query parameters can be done using an external library such as
github.com/go-playground/form.

```go
func NewQueryDecoder() httputil.Decoder {
	decoder := form.NewDecoder()
	decoder.SetTagName("url") // struct tag to use
	return httputil.DecodeFunc(func(r *http.Request, v interface{}) error {
		return decoder.Decode(v, r.URL.Query())
	})
}
```

Adopting the [`Decoder`][decoder] interface enables the usage of a common
validation middleware described above for both query parameters as well as the
request body.

### Encoding responses

[`JSONResponder`][jsonresponder] is a simple helper for responding to requests
with JSON either using a value or an error:

```go
func UserHandler(svc myapp.UserService) http.HandlerFunc {
	var responder httputil.JSONResponder
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "userID")
		user, err := svc.User(r.Context(), id)
		if err != nil {
			responder.Error(r, w, err)
			return
		}

		responder.Respond(r, w, myapp.Response{
			Success: true,
			Data:    user,
		})
	}
}
```

By default, when responding with a value, the status is set to `200: OK` but
this can be overridden using
[`JSONResponder.RespondWithStatus`][jsonresponder.respondwithstatus]:

```go
responder.RespondWithStatus(r, w, http.StatusCreated, myapp.Response{
	Success: true,
	Data:    id,
})
```

#### Encoding errors

[`JSONResponder`][jsonresponder] builds upon the interfaces declared in the
`github.com/sudo-suhas/xgo/errors` package to translate the `error` value into
the status and response body suitable to be sent to the caller.

[`JSONResponder.Error`][jsonresponder.error] leverages the
[`errors.StatusCoder`][errors.statuscoder] interface to infer the status code to
be set for sending the response.

```go
type StatusCoder interface {
	StatusCode() int
}
```

The status code for the error response can be overridden using
[`JSONResponder.ErrorWithStatus`][jsonresponder.errorwithstatus]:

```go
responder.ErrorWithStatus(r, w, http.StatusServiceUnavailable, err)
```

For transforming the error into the response body, a default implementation is
provided but it can also be overridden by specifying `ErrToRespBody` on the
[`JSONResponder`][jsonresponder] instance:

```go
var genericErrMsg = "We are not able to process your request. Please try again."

func newJSONResponder() httputil.JSONResponder {
	return httputil.JSONResponder{ErrToRespBody: errToRespBody}
}

func errToRespBody(err error) interface{} {
	// If the error does not implement xgo.JSONer interface, always return a
	// generic error response.
	var jsoner xgo.JSONer
	if !errors.As(err, &jsoner) {
		return myapp.GenericResponse{
			Errors: []myapp.ErrorResponse{{Message: genericErrMsg}},
		}
	}

	var errs []myapp.ErrorResponse
	// Extract the JSON representation of the error; Supported types -
	// myapp.ErrorResponse or a slice of myapp.ErrorResponse. Fallback to
	// generic error response for other types.
	switch v := jsoner.JSON().(type) {
	case myapp.ErrorResponse:
		errs = []myapp.ErrorResponse{v}

	case []myapp.ErrorResponse:
		errs = v

	default:
		errs = []myapp.ErrorResponse{{Message: genericErrMsg}}
	}

	return myapp.GenericResponse{Errors: errs}
}
```

To know more about `xgo.JSON` in the context of an `error`, see
[HTTP interop - Response Body](../errors#response-body) in the usage
documentation of the `errors` package.

By default, [`errors.UserMsg`][errors.usermsg] and [`xgo.JSONer`][xgo.jsoner]
are used to return a meaningful response to the caller. Example response JSON:

```go
msg := "The requested resource was not found."
responder.Error(r, w, errors.E(errors.WithOp("Get"), errors.NotFound, errors.WithUserMsg(msg)))
```

```json
{
	"success": false,
	"msg": "The requested resource was not found.",
	"errors": [
		{
			"code": "NOT_FOUND",
			"error": "not found",
			"msg": "The requested resource was not found."
		}
	]
}
```

#### Observing errors

Tracking errors, be it logging or instrumentation, is an important aspect and it
can be done easily by specifying `ErrObservers` on the
[`JSONResponder`][jsonresponder] instance:

```go
func newJSONResponder() httputil.JSONResponder {
	return httputil.JSONResponder{
		// Called for each error and can 'track' the error.
		ErrObservers: []httputil.ErrorObserverFunc{errLogger},
	}
}

func errLogger(r *http.Request, err error) {
	var e *errors.Error
	if !errors.As(err, &e) {
		httplog.LogEntrySetField(r, "error", err.Error())
		return
	}

	httplog.LogEntrySetField(r, "error_details", e.Details())
}
```

The observers are called by [`JSONResponder.Error`][jsonresponder.error] and
[`JSONResponder.ErrorWithStatus`][jsonresponder.errorwithstatus] for each error.
It is not recommended to do any time intensive operation inside the observer
functions as they are called synchronously in sequence.

### Building URLs

[`URLBuilder`][urlbuilder] makes building URLs convenient and prevents common
mistakes.

When calling an HTTP API, it typically involves combining dynamic parameters to
build the URL:

```go

const apiURL = "https://api.example.com/"

func userPostsURL(id, blogID string, limit, offset int) string {
	return fmt.Sprintf("%susers/%s/blogs/%s/posts?limit=%d&offset=%s", id, blogID, limit, offset)
}
```

Whenever using `apiURL`, we have to remember that it ends with a trailing slash
and ensure that we don't include the leading slash in the request URL path. And
more concerning is that the path parameters, `id` and `blogID`, are not escaped
appropriately.

We can use the `url` package to try and do this right:

```go
func userPostsURL(id, blogID string, limit, offset int) (*url.URL, error) {
	u, err := url.Parse(apiURL)
	if err != nil {
		return nil, err
	}

	p := fmt.Sprintf("/users/%s/blogs/%s", url.PathEscape(id), url.PathEscape(blogID))
	u.Path = path.Join(u.Path, p)

	q := u.Query()
	q.Set("limit", strconv.Itoa(limit))
	q.Set("offset", strconv.Itoa(offset))
	u.RawQuery = q.Encode()

	return u, nil
}
```

However, this might seem a bit tedious to write for each endpoint that we
integrate with. Additionally, we have to parse the `apiURL` each time since we
mutate the URL instance (copying is an alternative but _also_ tedious). This is
where [`URLBuilder`][urlbuilder] can help:

```go
type APIClient struct {
	httputil.URLBuilderSource

	http *http.Client
}

func NewAPIClient(apiURL string) (APIClient, error) {
	b, err = httputil.NewURLBuilderSource(apiURL)
	if err != nil {
		return nil, err
	}

	hc := http.Client{Timeout: 5 * time.Second}
	return APIClient{URLBuilderSource: b, http: &hc}, nil
}

func (c APIClient) UserPosts(ctx context.Context, id, blogID string, limit, offset int) ([]UserPost, error) {
	u := c.NewURLBuilder().
		Path("/users/{userID}/blogs/{blogID}").
		PathParam("userID", id).
		PathParam("blogID", blogID).
		QueryParamInt("limit", limit).
		QueryParamInt("offset", offset)

	r, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	// ...
}
```

[`URLBuilder`][urlbuilder] handles escaping the path parameters, encoding the
query parameters and building the complete URL.

**Examples:**

```go
b, err := httputil.NewURLBuilderSource("https://api.example.com/")
if err != nil {
	// ...
}

var u *url.URL
u = b.NewURLBuilder().
	Path("/users/{id}/posts").
	PathParamInt("id", 123).
	QueryParamInt("limit", 10).
	QueryParamInt("offset", 120).
	URL()
fmt.Println(u) // https://api.example.com/users/123/posts?limit=10&offset=120

u = b.NewURLBuilder().
	Path("/posts/{title}").
	PathParam("title", `Letters / "Special" Characters`).
	URL()
fmt.Println(u) // https://api.example.com/posts/Letters%2520%252F%2520%2522Special%2522%2520Characters

u = b.NewURLBuilder().
	Path("/users/{userID}/posts/{postID}/comments").
	PathParam("userID", "foo").
	PathParam("postID", "bar").
	QueryParams(url.Values{
		"search": {"some text"},
		"limit":  {"10"},
	}).
	URL()
fmt.Println(u) // https://api.example.com/users/foo/posts/bar/comments?limit=10&search=some+text
```

[pkg-go-dev-xgo-badge]: https://pkg.go.dev/badge/github.com/sudo-suhas/xgo
[pkg-go-dev-xgo-httputil]: https://pkg.go.dev/github.com/sudo-suhas/xgo/httputil
[decoder]: https://pkg.go.dev/github.com/sudo-suhas/xgo/httputil#Decoder
[jsondecoder]: https://pkg.go.dev/github.com/sudo-suhas/xgo/httputil#JSONDecoder
[validatingdecodermiddleware]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/httputil#ValidatingDecoderMiddleware
[validator]: https://pkg.go.dev/github.com/sudo-suhas/xgo#Validator
[jsonresponder]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/httputil#JSONResponder
[jsonresponder.respondwithstatus]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/httputil#JSONResponder.RespondWithStatus
[jsonresponder.error]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/httputil#JSONResponder.Error
[errors.statuscoder]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors#StatusCoder
[jsonresponder.errorwithstatus]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/httputil#JSONResponder.ErrorWithStatus
[errors.usermsg]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#UserMsg
[xgo.jsoner]: https://pkg.go.dev/github.com/sudo-suhas/xgo?tab=doc#JSONer
[urlbuilder]: https://pkg.go.dev/github.com/sudo-suhas/xgo/httputil#URLBuilder
