# xgo [![build][github-workflow-badge]][github-workflow] [![PkgGoDev][pkg-go-dev-xgo-badge]][pkg-go-dev-xgo] [![Go Report Card][go-report-card-badge]][go-report-card]

> Collection of packages to serve as the building blocks for developing Go
> applications.

## Install

```
$ go get github.com/sudo-suhas/xgo
```

## Usage

Usage for each package is documentated in the respective readme.

- [`errors`](errors#table-of-contents) ([API reference][errors-api-docs])
- [`httputil`](httputil#table-of-contents) ([API reference][httputil-api-docs])

## Decision Log

The rationale for important design decisions is documented in
[decision-log.md](decision-log.md).

## Credits

A lot of the ideas for the errors package are from these articles and talks:

- ["Error handling in Upspin"][err-handling-upspin] by Rob Pike and Andrew
  Gerrand.
- ["Failure is your Domain"][failure-your-domain] by Ben Johnson.
- ["Handling Go errors"][handling-go-errors] by Marwan Sulaiman from
  GopherCon 2019.

The JSON decoder implementation in the `httputil` package is based on ["How to
Parse a JSON Request Body in Go"][how-to-parse-json-req] by Alex Edwards.

The URL builder in the `httputil` package drew some inspiration from
https://github.com/balazsbotond/urlcat.

[github-workflow-badge]:
	https://github.com/sudo-suhas/xgo/workflows/build/badge.svg
[github-workflow]:
	https://github.com/sudo-suhas/xgo/actions?query=workflow%3Abuild
[pkg-go-dev-xgo-badge]: https://pkg.go.dev/badge/github.com/sudo-suhas/xgo
[pkg-go-dev-xgo]: https://pkg.go.dev/mod/github.com/sudo-suhas/xgo?tab=packages
[go-report-card-badge]: https://goreportcard.com/badge/github.com/sudo-suhas/xgo
[go-report-card]: https://goreportcard.com/report/github.com/sudo-suhas/xgo
[errors-api-docs]: https://pkg.go.dev/github.com/sudo-suhas/xgo/errors
[httputil-api-docs]: https://pkg.go.dev/github.com/sudo-suhas/xgo/httputil
[err-handling-upspin]:
	https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html
[failure-your-domain]: https://middlemost.com/failure-is-your-domain/
[handling-go-errors]:
	https://about.sourcegraph.com/go/gophercon-2019-handling-go-errors
[how-to-parse-json-req]:
	https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body
