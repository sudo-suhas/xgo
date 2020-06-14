package errors_test

import (
	"fmt"

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

func ExampleMatch() {
	msg := "Oops! Something went wrong. Please try again after some time."
	err := errors.New("network unreachable")

	// Construct an error, one we pretend to have received from a test.
	got := errors.E(
		errors.WithOp("Get"),
		errors.WithUserMsg(msg),
		errors.IO,
		errors.WithErr(err),
	)

	// Now construct a reference error, which might not have all
	// the fields of the error from the test.
	expect := errors.E(
		errors.WithOp("Get"),
		errors.IO,
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
