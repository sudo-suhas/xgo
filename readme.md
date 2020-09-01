# xgo ![build](https://github.com/sudo-suhas/xgo/workflows/build/badge.svg) [![PkgGoDev](https://pkg.go.dev/badge/github.com/sudo-suhas/xgo)](https://pkg.go.dev/mod/github.com/sudo-suhas/xgo?tab=packages)

> Collection of packages to serve as the building blocks for developing Go
> applications.

## Install

```
$ go get github.com/sudo-suhas/xgo
```

## Usage

Usage for each package is documentated in the respective readme.

- [`errors`](errors#table-of-contents) ([API reference][errors-api-docs])
- [`httputil`](httputil) ([API reference][httputil-api-docs])

## Decision Log

The rationale for important design decisions is documented in
[decision-log.md](decision-log.md).

## Credit

A lot of the ideas for the errors package are from these articles and talks:

- ["Error handling in Upspin"][err-handling-upspin] by Rob Pike and Andrew
  Gerrand.
- ["Failure is your Domain"][failure-your-domain] by Ben Johnson.
- ["Handling Go errors"][handling-go-errors] by Marwan Sulaiman from
  GopherCon 2019.

[errors-api-docs]: https://pkg.go.dev/github.com/sudo-suhas/xgo/errors
[httputil-api-docs]: https://pkg.go.dev/github.com/sudo-suhas/xgo/httputil
[err-handling-upspin]:
	https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html
[failure-your-domain]: https://middlemost.com/failure-is-your-domain/
[handling-go-errors]:
	https://about.sourcegraph.com/go/gophercon-2019-handling-go-errors
