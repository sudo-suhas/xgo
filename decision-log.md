# Decision Log

> Log of design decisions and the rationale.

## Table of Contents

- [`errors.E(...)`](#errorse)
- [`errors.Kind`](#errorskind)
- [`xgo.JSONer` vs `json.Marshaler`](#xgojsoner-vs-jsonmarshaler)

### `errors.E(...)`

The `errors` package draws inspiration from the approach described in ["Error
handling in Upspin"][err-handling-upspin] by Rob Pike and Andrew Gerrand.

The constructor used in `upspin.io/errors` uses variadic arguments of
`interface{}` type:

```go
func E(args ...interface{}) error
```

The goal was to make constructing an error as easy as possible with only the
arguments which are relevant in the specific context. And the constructor uses
the `type` of each argument to infer the field it is meant for.

In practice, I found the following limitations:

- The function signature is very lenient. The lack of compile-time safety can
  lead to inadvertent developer mistakes such as passing an argument of an
  unsupported data type.
- Restriction on the data types for fields in the error struct. Since it
  wouldn't be possible to determine which field the argument was meant for if
  there were two fields with the same data type.

Ben Johnson, in his blog post ["Failure is your Domain"][failure-your-domain],
builds upon the approach in `upspin.io/errors` with a few differences. Errors
are created explicitly without the use of a constructor by directly supplying
the relevant fields. Example:
`&myapp.Error{Code: myapp.EINVALID, Message: "Username is required."}`. The
reason for _this_ `errors` package to not take this approach is that having a
constructor makes it possible to promote fields from the underlying error (if it
is of the type `errors.Error`).

In summary, the design of the constructor had the following goals:

- Of the multiple fields that the `Error` has, it should be possible to only
  supply the fields which are relevant to the specific context.
- Compile-time safety to guard against developer mistakes.

The best answer I could find was to use functional options. It is no doubt more
verbose than the `upspin.io/errors` approach but the trade-off between verbosity
and clarity is not too bad (I hope).

```go
const op xgo.Op = "service.MakeBooking"
//...
if err := db.OpenConn(ctx); err != nil {
	return nil, errors.E(errors.WithOp(op), errors.IO, errors.WithText("open connection"), errors.WithErr(err))
}
```

Using the functional options pattern has an added advantage - extensibility. It
is trivial to write an `errors.Option` which can hook into the construction of
the error instance.

```go
func withSQLError(err error) errors.Option {
	return errors.OptionFunc(func(e *errors.Error) {
		if err == sql.ErrNoRows {
			err.Kind = errors.NotFound
		}

		e.Err = err
	})
}
```

A similar approach can also be used to interpret the error response from a
service that your application depends on.

### `errors.Kind`

Datatypes considered for `errors.Kind`:

- [Go enum](#go-enum)
- [String](#string)
- [Integer](#integer)
- [Struct/tuple](#struct-tuple) (**Winner**)

#### Go enum

> Ref: https://pkg.go.dev/upspin.io/errors?tab=doc#Kind

Why not go with the approach used in `upspin.io/errors`? The most important
reason is that the `errors` package would no longer be general-purpose as
defining new `Kind`s in the application domain wouldn't be feasible. We would
have methods `String(), ErrorCode()` which would not be aware of any `Kind`
defined outside of the `errors` package.

#### String

> Ref: https://middlemost.com/failure-is-your-domain/#defining-our-error-codes

The simplicity of a string value was hard to pass up. For the most part, this
data type works. But the problem was the implicit association of `Kind` with
HTTP status code which would have to be pre-defined in the `errors` package
making it difficult to map a status code to a custom `Kind` defined in the
application domain.

#### Integer

> Ref: https://www.w3.org/Protocols/rfc2616/rfc2616-sec10.html

One of the core proposition of the `errors` package is being able to associate
the HTTP response code with the type of error. To accomplish this, without the
use of an enum, the value itself could be the generic status codes defined by
HTTP. However, this has limited flexibility. We would not be able to define
different error `Kind`s with the same HTTP status codes. Some status codes such
as `431: Request Header Fields Too Large` are _very_ specific and wouldn't be
useful for the vast majority of scenarios.

#### Struct/tuple

**_This is the selected approach._**

Go structs with a few fields are very lightweight and can be considered very
similar to a tuple. And having a tuple of the error code and status code makes
the association explicit along with the freedom to define a custom `Kind` in the
application domain. This has the downside of having to use an attribute which
might not always be relevant for the application. But if the status code is not
relevant, it can be omitted when declaring any custom `Kind`. And since the
`Kind` struct is comparable, both for equality and switch statements, it can be
used as though it is just a constant.

```go
if errors.WhatKind(err) == errors.NotFound {
	// ...
}
```

### `xgo.JSONer` vs `json.Marshaler`

`xgo.JSONer`:

```go
type JSONer interface {
	JSON() interface{}
}
```

The reason for defining, and implementing this interface in `*errors.Error` is
to provide greater control over how the `JSON` representation of an `error` is
used. The `error` is free to return the most appropriate `JSON` representation
and leave it up to the caller on how to use it.

Here's some example code which uses reflection to identify the underlying type
and combine the JSON representation of multiple errors:

```go
func errsToJSON(errs ...error) []interface{} {
	var jj []interface{}
	for _, err := range errs {
		ej, ok := err.(xgo.JSONer)
		if !ok {
			continue
		}

		switch j := ej.JSON(); reflect.TypeOf(j).Kind() {
		case reflect.Slice:
			s := reflect.ValueOf(j)
			for i := 0; i < s.Len(); i++ {
				jj = append(jj, s.Index(i).Interface())
			}

		case reflect.Map, reflect.Struct:
			jj = append(jj, j)
		}
	}

	return jj
}
```

Similarly, in the context of a web application, when the request results in an
error, to be able to respond with a consistent structure, we can check if the
value returned from `*errors.Error.JSON()` is a single value or a slice. If it
is a slice, we can directly pass the result to the `errors` field. Otherwise, we
can wrap it in a slice. Example JSON response:

```json
{
	"success": false,
	"errors": [
		{
			"code": "NOT_FOUND",
			"error": "not found",
			"msg": "The requested resource was not found."
		}
	]
}
```

[err-handling-upspin]:
	https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html
[failure-your-domain]: https://middlemost.com/failure-is-your-domain/
