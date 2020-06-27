# errors [![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/github.com/sudo-suhas/xgo/errors)

General purpose error handling package with few extra bells and whistles for
HTTP interop.

## Table of contents

- [Usage](#usage)
  - [Creating errors](#creating-errors)
  - [Adding context to an error](#adding-context-to-an-error)
  - [Inspecting errors](#inspecting-errors)
  - [Errors for the end user](#errors-for-the-end-user)
  - [HTTP interop](#http-interop)
    - [Status code](#status-code)
    - [Response body](#response-body)
  - [Logging errors](#logging-errors)
- [Errors package objectives](#errors-package-objectives)

## Usage

### Creating errors

To facilitate error construction, the package provides a function,
[`errors.E`][errors.e] which build an [`*Error`][errors.error] from its
(functional) options:

```go
func E(opt Option, opts ...Option) error
```

In typical use, calls to [`errors.E`][errors.e] will arise multiple times within
a method, so a constant called `op` could be defined that will be passed to all
`E` calls within the method:

```go
func (b *binder) Bind(r *http.Request, v interface{}) error {
	const op xgo.Op = "binder.Bind"

	if err := b.Decode(r, v); err != nil {
		return errors.E(errors.WithOp(op), errors.InvalidInput, errors.WithErr(err))
	}

	if err := b.Validate.Struct(v); err != nil {
		return errors.E(errors.WithOp(op), errors.InvalidInput, errors.WithErr(err))
	}

	return nil
}
```

The error string, for the example above, would look something like this:

```
binder.Bind: invalid input: json: cannot unmarshal number into Go struct field .Name of type string
```

### Adding context to an error

A new error wrapping the original error is returned by passing it in to the
constructor. Additional context can include the following (optional):

- [`errors.WithOp(xgo.Op)`][errors.withop]: Operation name being attempted.
- [`errors.Kind`][errors.kind]: Error classification.
- [`errors.WithText(string)`][errors.withtext]: Error string.
  [`errors.WithTextf(string, ...interface{})`][errors.withtextf] can also be
  used to format the error string with additional arguments.
- [`errors.WithData(interface{})`][errors.withdata]: Arbitrary value which could
  be considered relevant to the error.

```go
if err := svc.SaveOrder(o); err != nil {
	return errors.E(errors.WithOp(op), errors.WithErr(err))
}
```

### Inspecting errors

The error [`Kind`][errors.kind] can be extracted using the function
[`errors.WhatKind(error)`][errors.whatkind]:

```go
if errors.WhatKind(err) == errors.NotFound {
	// ...
}
```

With this, it is straightforward for the app to handle the error appropriately
depending on the classification, such as a permission error or a timeout error.

There is also [`errors.Match`][errors.match] which can be useful in tests to
compare and check only the properties which are of interest. This allows to
easily ignore the irrelevant details of the error.

```go
func Match(template, err error) bool
```

The function checks whether the error is of type
[`*errors.Error`][errors.error], and if so, whether the fields within equal
those within the template. The key is that it checks only those fields that are
non-zero in the template, ignoring the rest.

```go
if errors.Match(errors.E(errors.WithOp("service.MakeBooking"), errors.PermissionDenied), err) {
	// ...
}
```

The functions [`errors.Is`][errors.is] and [`errors.As`][errors.as], which were
[introduced in Go 1.13][go1.13-errors], can also be used. These functions are
defined in the package for the sake of convenience and delegate to the standard
library. This way, importing both this package and the `errors` package from the
standard library should not be necessary and avoids having to work around the
name conflict on importing both.

```go
// For an error returned from the the database client
// 	return errors.E(errors.WithOp(op), errors.WithErr(err))
if errors.Is(err, sql.ErrNoRows) {
	// ...
}

// In the API client
var terr interface{ Timeout() bool }
if errors.As(err, &terr) && terr.Timeout() {
	// ...
}
```

### Errors for the end user

Errors have multiple consumers, the end user being one of them. However, to the
end user, the error text, root cause etc. are not relevant (and in most cases,
**should not** be exposed as it could be a security concern).

To address this, [`Error`][errors.error] has the field `UserMsg` which is
intended to be returned/shown to the user.
[`errors.WithUserMsg(string)`][errors.withusermsg] stores the message into the
aforementioned field. And it can be retrieved using
[`errors.UserMsg`][errors.usermsg]

```go
func UserMsg(err error) string
```

Example:

```go
// CreateUser creates a new user in the system.
func (s *Service) CreateUser(ctx context.Context, user *myapp.User) error {
	const op xgo.Op = "svc.CreateUseer"

	// Validate username is non-blank.
	if user.Username == "" {
		msg := "Username is required"
		return errors.E(errors.WithOp(op), errors.InvalidInput, errors.WithUserMsg(msg))
	}

	// Verify user does not already exist
	if s.usernameInUse(user.Username) {
		msg := "Username is already in use. Please choose a different username."
		return errors.E(errors.WithOp(op), errors.AlreadyExists, errors.WithUserMsg(msg))
	}

	// ...
}

// Elsewhere in the application, for responding with the error to the user
if msg := errors.UserMsg(err); msg != "" {
	// ...
}
```

Translating the error to a meaningful HTTP response is convered in the section
[HTTP interop - Response body](#response-body).

### HTTP interop

- [Status code](#status-code)
- [Response body](#response-body)

#### Status code

`errors` is intended to be a general purpose error handling package. However, it
includes minor extensions to make working with HTTP easier. The idea is that its
there if you need it but doesn't get in the way if you don't.

[`errors.Kind`][errors.kind] is composed of the error code and the HTTP status
code.

```go
type Kind struct {
	Code   string
	Status int
}
```

_Rationale for the design of [`Kind`][errors.kind] is documented in
[decision-log.md](/decision-log.md#errorskind)_

The [`Kind`][errors.kind] variables [defined in the `errors`
package][errors.pkg-variables] include mapping to the appropriate HTTP status
code. For example, `InvalidInput` maps to `400: Bad Request` while `NotFound`
maps to `404: Not Found`.

Therefore, provided an error has an appropriate [`Kind`][errors.kind]
associated, it is really simple to translate the same to the HTTP status code
using [`errors.StatusCode`][errors.statuscode]

```go
func StatusCode(err error) int
```

Conversely, [`errors.KindFromStatus`][errors.kindfromstatus] can be used to
translate the HTTP response code from an external REST API to the
[`Kind`][errors.kind].

```go
func KindFromStatus(status int) Kind
```

_The utility function [`errors.WithResp(*http.Response)`][errors.withresp],
which uses [`KindFromStatus`][errors.kindfromstatus], makes it possible to
construct an error from `*http.Response` with contextual information of the
request and response._

#### Response body

Another concern of web applications is returning a meaningful response in case
there is an error. [`*Error`][errors.error] implements the
[`xgo.JSONer`][xgo.jsoner] interface which can be leveraged for building the
error response.

```go
var ej xgo.JSONer
if errors.As(err, &ej) {
	j := ej.JSON()
	// ...
}
```

For example, consider the following error:

```go
msg := "The requested resource was not found."
errors.E(errors.WithOp("Get"), errors.NotFound, errors.WithUserMsg(msg))
```

This produces the following JSON representation:

```json
{
	"code": "NOT_FOUND",
	"error": "not found",
	"msg": "The requested resource was not found."
}
```

If needed, this behaviour can easily be tweaked by using
[`errors.WithToJSON(errors.JSONFunc)`][errors.withtojson]. It expects a function
of type [`errors.JSONFunc`][errors.jsonfunc]

```go
type JSONFunc func(*Error) interface{}
```

Simple example illustrating the usage of
[`errors.WithToJSON`][errors.withtojson]:

```go
func makeErrToJSON(lang string) errors.JSONFunc {
	return func(e *errors.Error) interface{} {
		msgKey := errors.UserMsg(e)
		return map[string]interface{}{
			"code":    errors.WhatKind(e).Code,
			"title":   i18n.Title(lang, msgKey),
			"message": i18n.Message(lang, msgKey),
		}
	}
}

// In the HTTP handler, when there is an error, wrap it and pass it to the
// response formatter.
errors.E(errors.WithToJSON(makeErrToJSON(lang)), errors.WithErr(err))
```

_It is recommended to pass in the `JSONFunc` to the error only **once**; i.e the
error should not be wrapped repeatedly with virtually the same function passed
into [`errors.WithToJSON`][errors.withtojson]. The reason is that functions are
not comparable in Go and hence cannot be de-duplicated for keeping the error
chain as short as possible. The obvious exception is when we want to override
any `ToJSON` which might have been supplied in calls to [`errors.E`][errors.e]
further down the stack_

### Logging errors

Logs are meant for a developer or an operations person. Such a person
understands the details of the system and would benefit from being to see as
much information related to the error as possible. This includes the _logical
stack trace_ (to help understand the program flow), error code, error text and
any other contextual information associated with the error.

The error string by default includes most of the information but would not be
ideal for parsing by machines and is therefore less friendly to structured
search.

[`*Error.Details()`][error.details] constructs and yields the details of the
error, by traversing the error chain, in a [structure][errors.internaldetails]
which is suitable to be parsed.

```go
type InternalDetails struct {
	Ops   []xgo.Op    `json:"ops,omitempty"`
	Kind  Kind        `json:"kind,omitempty"`
	Error string      `json:"error"`
	Data  interface{} `json:"data,omitempty"`
}
```

_With respect to stack traces, dumping a list of every function in the call
stack from where an error occurred can many times be overwhelming. After all,
for understanding the context of an error, only a small subset of those lines is
needed. A logical stack trace contains only the layers that we as developers
find to be important in describing the program flow._

## Errors package objectives

- Composability: Being able to compose an error using a different error and
  optionally associate additional contextual information/data.
- Classification: Association of error type/classification and being able to
  test for the same.
- Flexibility: Create errors with only the information deemed necessary,
  _almost_ everything is optional.
- Interoperability:
  - HTTP: Map errors classification to HTTP status code to encompass the
    response code in the error.
  - Traits: Define and use traits such as `GetKind`, `StatusCoder` to allow
    interoperability with custom error definitions.
- Descriptive: Facilitate reporting the sequence of events, with relevant
  contextual information, that resulted in a problem.
  > Carefully constructed errors that thread through the operations in the
  > system can be more concise, more descriptive, and more helpful than a simple
  > stack trace.

[errors.e]: https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#E
[errors.error]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#Error
[errors.withop]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#WithOp
[errors.kind]: https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#Kind
[errors.withtext]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#WithText
[errors.withtextf]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#WithTextf
[errors.withdata]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#WithData
[errors.whatkind]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#WhatKind
[errors.is]: https://pkg.go.dev/errors?tab=doc#Is
[errors.as]: https://pkg.go.dev/errors?tab=doc#As
[go1.13-errors]: https://blog.golang.org/go1.13-errors
[errors.match]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#Match
[errors.withusermsg]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#WithUserMsg
[errors.usermsg]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#UserMsg
[errors.pkg-variables]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#pkg-variables
[errors.statuscode]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#StatusCode
[errors.kindfromstatus]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#KindFromStatus
[errors.withresp]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#WithResp
[xgo.jsoner]: https://pkg.go.dev/github.com/sudo-suhas/xgo?tab=doc#JSONer
[errors.withtojson]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#WithToJSON
[errors.jsonfunc]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#JSONFunc
[error.details]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#Error.Details
[errors.internaldetails]:
	https://pkg.go.dev/github.com/sudo-suhas/xgo/errors?tab=doc#InternalDetails
