package errors

import "net/http"

type MyError struct{}

func (MyError) Error() string   { return "e" }
func (MyError) GetKind() Kind   { return Internal }
func (MyError) StatusCode() int { return http.StatusForbidden }
