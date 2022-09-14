// Package errors is a general purpose error handling package, with few
// extra bells and whistles for HTTP interop.
//
// # Creating errors
//
// To facilitate error construction, the package provides a function,
// errors.E which build an *Error from its (functional) options.
//
//	func E(opt Option, opts ...Option) error
//
// In typical use, calls to errors.E will arise multiple times within a
// method, so a constant called op could be defined that will be passed
// to all E calls within the method:
//
//	func (b *binder) Bind(r *http.Request, v interface{}) error {
//		const op = "binder.Bind"
//
//		if err := b.Decode(r, v); err != nil {
//			return errors.E(errors.WithOp(op), errors.InvalidInput, errors.WithErr(err))
//		}
//
//		if err := b.Validate.Struct(v); err != nil {
//			return errors.E(errors.WithOp(op), errors.InvalidInput, errors.WithErr(err))
//		}
//
//		return nil
//	}
//
// # Adding context to an error
//
// A new error wrapping the original error is returned by passing it in
// to the constructor. Additional context can include the following
// (optional):
//
//   - errors.WithOp(string): Operation name being attempted.
//   - errors.Kind: Error classification.
//   - errors.WithText(string): Error string.
//     errors.WithTextf(string, ...interface{}) can also be
//     used to format the error string with additional arguments.
//   - errors.WithData(interface{}): Arbitrary value which could be
//     considered relevant to the error.
//
// Example:
//
//	if err := svc.SaveOrder(o); err != nil {
//		return errors.E(errors.WithOp(op), errors.WithErr(err))
//	}
//
// # Inspecting errors
//
// The error Kind can be extracted using the function WhatKind(error):
//
//	if errors.WhatKind(err) == errors.NotFound {
//		// ...
//	}
//
// With this, it is straightforward for the app to handle the error
// appropriately depending on the classification, such as a permission
// error or a timeout error.
//
// There is also Match which can be useful in tests to compare and check
// only the properties which are of interest. This allows to easily
// ignore the irrelevant details of the error.
//
// The function checks whether the error is of type *Error, and if so,
// whether the fields within equal those within the template. The key is
// that it checks only those fields that are non-zero in the template,
// ignoring the rest.
//
//	if errors.Match(errors.E(errors.WithOp("service.MakeBooking"), errors.PermissionDenied), err) {
//		// ...
//	}
//
// # Errors for the end user
//
// Errors have multiple consumers, the end user being one of them.
// However, to the end user, the error text, root cause etc. are not
// relevant (and in most cases, should not be exposed as it could be a
// security concern).
//
// To address this, Error has the field UserMsg which is  intended to be
// returned/shown to the user. WithUserMsg(string) stores the message
// into the aforementioned field. And it can be retrieved using
// errors.UserMsg.
//
// Example:
//
//	// CreateUser creates a new user in the system.
//	func (s *Service) CreateUser(ctx context.Context, user *myapp.User) error {
//		const op = "svc.CreateUseer"
//
//		// Validate username is non-blank.
//		if user.Username == "" {
//			msg := "Username is required"
//			return errors.E(errors.WithOp(op), errors.InvalidInput, errors.WithUserMsg(msg))
//		}
//
//		// Verify user does not already exist
//		if s.usernameInUse(user.Username) {
//			msg := "Username is already in use. Please choose a different username."
//			return errors.E(errors.WithOp(op), errors.AlreadyExists, errors.WithUserMsg(msg))
//		}
//
//		// ...
//	}
//
//	// Elsewhere in the application, for responding with the error to the user
//	if msg := errors.UserMsg(err); msg != "" {
//		// ...
//	}
package errors
