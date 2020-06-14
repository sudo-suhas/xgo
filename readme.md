# xgo ![Build](https://github.com/sudo-suhas/xgo/workflows/Check/badge.svg) [![go.dev reference](https://img.shields.io/badge/go.dev-reference-007d9c?logo=go&logoColor=white&style=flat-square)](https://pkg.go.dev/mod/github.com/sudo-suhas/xgo?tab=packages)

> Collection of packages to serve as the building blocks for developing Go
> applications.

## Install

```
$ go get github.com/sudo-suhas/xgo
```

## Usage

TODO

## Decision Log

The rationale for important design decisions is documented in
[decision-log.md](/decision-log.md).

## Credit

A lot of the ideas for the errors package are from these articles and talks:

- ["Error handling in Upspin"][err-handling-upspin] by Rob Pike and Andrew
  Gerrand.
- ["Failure is your Domain"][failure-your-domain] by Ben Johnson.
- ["Handling Go errors"][handling-go-errors] by Marwan Sulaiman from
  GopherCon 2019.

[err-handling-upspin]:
	https://commandcenter.blogspot.com/2017/12/error-handling-in-upspin.html
[failure-your-domain]: https://middlemost.com/failure-is-your-domain/
[handling-go-errors]:
	https://about.sourcegraph.com/go/gophercon-2019-handling-go-errors