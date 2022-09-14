package errors_test

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/sudo-suhas/xgo/errors"
)

func ExampleE() {
	// Simple error
	e1 := errors.E(
		errors.WithOp("xgo_test.SimpleError"),
		errors.Internal,
		errors.WithText("fail"),
	)
	fmt.Println("\nSimple error:")
	fmt.Println(e1)

	// Nested error.
	e2 := errors.E(errors.WithOp("xgo_test.NestedError"), errors.WithErr(e1))
	fmt.Println("\nNested error:")
	fmt.Println(e2)

	// Output:
	// Simple error:
	// xgo_test.SimpleError: internal error: fail
	//
	// Nested error:
	// xgo_test.NestedError: internal error: xgo_test.SimpleError: fail
}

func ExampleWhatKind() {
	// Somewhere in the application, return the error with Kind: NotFound.
	err := errors.E(errors.WithOp("Get"), errors.NotFound, errors.WithErr(sql.ErrNoRows))

	// Use WhatKind to extract the Kind associated with the error.
	fmt.Println("Kind:", errors.WhatKind(err))

	// If required, define and use a custom Kind
	k := errors.Kind{Code: "CONFLICT"}
	err = errors.E(errors.WithOp("UpdateEntity"), k)

	fmt.Println("Custom kind:", errors.WhatKind(err))
	// Output:
	// Kind: not found
	// Custom kind: conflict
}

func ExampleStatusCode() {
	// Somewhere in the application, return the error with Kind: NotFound.
	err := errors.E(errors.WithOp("Get"), errors.NotFound, errors.WithErr(sql.ErrNoRows))

	// Use StatusCode to extract the status code associated with the error Kind.
	fmt.Println("Status code:", errors.StatusCode(err))

	// If required, define and use a custom Kind
	k := errors.Kind{Code: "CONFLICT", Status: http.StatusConflict}
	err = errors.E(errors.WithOp("UpdateEntity"), k)

	fmt.Println("Status code with custom kind:", errors.StatusCode(err))
	// Output:
	// Status code: 404
	// Status code with custom kind: 409
}

func ExampleUserMsg() {
	// Along with the error Kind, associate the appropriate message for the
	// specific error scenario.
	msg := "Username is required"
	e1 := errors.E(errors.WithOp("svc.CreateUser"), errors.InvalidInput, errors.WithUserMsg(msg))

	// Use UserMsg to extract the message to be shown to the user.
	fmt.Println("User message:", errors.UserMsg(e1))

	// Override the message by wrapping it in a new error with a different
	// message.
	e2 := errors.E(errors.WithErr(e1), errors.WithUserMsg("Deal with it!"))
	fmt.Println("Overidden user message:", errors.UserMsg(e2))
	// Output:
	// User message: Username is required
	// Overidden user message: Deal with it!
}

func ExampleMatch() {
	msg := "Oops! Something went wrong. Please try again after some time."
	err := errors.New("network unreachable")

	// Construct an error, one we pretend to have received from a test.
	got := errors.E(
		errors.WithOp("Get"),
		errors.WithUserMsg(msg),
		errors.Unavailable,
		errors.WithErr(err),
	)

	// Now construct a reference error, which might not have all
	// the fields of the error from the test.
	expect := errors.E(
		errors.WithOp("Get"),
		errors.Unavailable,
		errors.WithErr(err),
	)

	fmt.Println("Match:", errors.Match(expect, got))

	// Now one that's incorrect - wrong Kind.
	got = errors.E(
		errors.WithOp("Get"),
		errors.WithUserMsg(msg),
		errors.PermissionDenied,
		errors.WithErr(err),
	)

	fmt.Println("Mismatch:", errors.Match(expect, got))

	// Output:
	//
	// Match: true
	// Mismatch: false
}
