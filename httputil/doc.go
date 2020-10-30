// Package httputil provides HTTP utility functions, focused around
// decoding and responding with JSON.
//
// Decoding requests
//
// The httputil package defines the Decoder interface to serve as the
// central building block:
//
// 	type Decoder interface {
// 		// Decode decodes the HTTP request into the given value.
// 		Decode(r *http.Request, v interface{}) error
// 	}
//
// JSONDecoder implements this interface and can be used to parse the
// request body if the content type is JSON.
//
// Validation of input can be plugged into the decoding step using
// ValidatingDecoderMiddleware with an implementation of xgo.Validator:
//
// 	var vd xgo.Validator = MyValidator{}
// 	var dec httputil.Decoder
// 	{
// 		dec = httputil.JSONDecoder{}
// 		dec = httputil.ValidatingDecoderMiddleware(vd)(dec)
// 	}
//
// Encoding responses
//
// JSONResponder is a simple helper for responding to requests with JSON
// either using a value or an error:
//
// 	var responder httputil.JSONResponder
// 	// ...
// 	responder.Respond(r, w, myapp.Response{
// 		Success: true,
// 		Data:    result,
// 	})
//
// By default, when responding with a value, the status is set to
// '200: OK' but this can be overridden using
// JSONResponder.RespondWithStatus:
//
// 	responder.RespondWithStatus(r, w, http.StatusCreated, myapp.Response{
// 		Success: true,
// 		Data:    id,
// 	})
//
// JSONResponder builds upon the interfaces declared in the
// github.com/sudo-suhas/xgo/errors package to translate the error value
// into the status and response body suitable to be sent to the caller.
//
// JSONResponder.Error leverages the errors.StatusCoder interface to
// infer the status code to be set for sending the response.
//
// 	type StatusCoder interface {
// 		StatusCode() int
// 	}
//
// The status code for the error response can be overridden using
// JSONResponder.ErrorWithStatus:
//
// 	responder.ErrorWithStatus(r, w, http.StatusServiceUnavailable, err)
//
// For transforming the error into the response body, a default
// implementation is provided but it can also be overridden by
// specifying ErrToRespBody on the JSONResponder instance:
//
// 	var genericErrMsg = "We are not able to process your request. Please try again."
//
// 	func newJSONResponder() httputil.JSONResponder {
// 		return httputil.JSONResponder{ErrToRespBody: errToRespBody}
// 	}
//
// 	func errToRespBody(err error) interface{} {
// 		// A contrived implementation of the transform func.
// 		return myapp.GenericResponse{
// 			Errors: []myapp.ErrorResponse{{Message: genericErrMsg}},
// 		}
// 	}
//
// Observing errors
//
// Tracking errors, be it logging or instrumentation, is an important
// aspect and it can be done easily by specifying ErrObservers on the
// JSONResponder instance:
//
// 	func newJSONResponder() httputil.JSONResponder {
// 		return httputil.JSONResponder{
// 			// Called for each error and can 'track' the error.
// 			ErrObservers: []httputil.ErrorObserverFunc{errLogger},
// 		}
// 	}
//
// 	func errLogger(r *http.Request, err error) {
// 		var e *errors.Error
// 		if !errors.As(err, &e) {
// 			httplog.LogEntrySetField(r, "error", err.Error())
// 			return
// 		}
//
// 		httplog.LogEntrySetField(r, "error_details", e.Details())
// 	}
//
package httputil
