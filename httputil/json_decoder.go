package httputil

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"

	"github.com/sudo-suhas/xgo"
	"github.com/sudo-suhas/xgo/errors"
)

var (
	ErrKindUnsupportedMediaType = errors.Kind{
		Code:   "UNSUPPORTED_MEDIA_TYPE",
		Status: http.StatusUnsupportedMediaType,
	}
	ErrKindRequestEntityTooLarge = errors.Kind{
		Code:   "REQUEST_ENTITY_TOO_LARGE",
		Status: http.StatusRequestEntityTooLarge,
	}
)

// JSONDecoder decodes the request body into the given value. It expects
// the request body to be JSON.
type JSONDecoder struct {
	// SkipCheckContentType, if set to true, skips the check on
	// value of Content-Type header being "application/json".
	SkipCheckContentType bool

	// UseNumber causes the Decoder to unmarshal a number into an
	// interface{} as a Number instead of as a float64.
	UseNumber bool

	// DisallowUnknownFields causes the Decoder to return an error when
	// the destination is a struct and the input contains object keys
	// which do not match any non-ignored, exported fields in the
	// destination.
	DisallowUnknownFields bool
}

func (j JSONDecoder) Decode(r *http.Request, v interface{}) error {
	var op xgo.Op = "JSONDecoder.Decode"

	defer io.Copy(ioutil.Discard, r.Body) //nolint:errcheck

	// Based on https://www.alexedwards.net/blog/how-to-properly-parse-a-json-request-body

	if err := j.checkContentType(r); err != nil {
		return errors.E(errors.WithOp(op), errors.WithErr(err))
	}

	dec := j.newDecoder(r.Body)
	if err := dec.Decode(v); err != nil {
		var (
			syntaxErr *json.SyntaxError
			typeErr   *json.UnmarshalTypeError
		)
		switch {
		// Catch any syntax errors in the JSON and send an error message
		// which interpolates the location of the problem to make it
		// easier for the client to fix.
		case errors.As(err, &syntaxErr):
			msg := fmt.Sprintf("Request body contains badly-formed JSON (at position %d)", syntaxErr.Offset)
			return errors.E(
				errors.WithOp(op), errors.InvalidInput, errors.WithUserMsg(msg), errors.WithErr(err),
			)

		// In some circumstances Decode() may also return an
		// io.ErrUnexpectedEOF error for syntax errors in the JSON. There
		// is an open issue regarding this at
		// https://github.com/golang/go/issues/25956.
		case errors.Is(err, io.ErrUnexpectedEOF):
			msg := "Request body contains badly-formed JSON"
			return errors.E(
				errors.WithOp(op), errors.InvalidInput, errors.WithUserMsg(msg), errors.WithErr(err),
			)

		// Catch any type errors, like trying to assign a string in the
		// JSON request body to a int field in our Person struct. We can
		// interpolate the relevant field name and position into the error
		// message to make it easier for the client to fix.
		case errors.As(err, &typeErr):
			msg := fmt.Sprintf(
				"Request body contains an invalid value for the '%s' field (at position %d)",
				typeErr.Field, typeErr.Offset,
			)
			return errors.E(
				errors.WithOp(op), errors.InvalidInput, errors.WithUserMsg(msg), errors.WithErr(err),
			)

		// Catch the error caused by extra unexpected fields in the request
		// body. We extract the field name from the error message and
		// interpolate it in our custom error message. There is an open
		// issue at https://github.com/golang/go/issues/29035 regarding
		// turning this into a sentinel error.
		case strings.HasPrefix(err.Error(), "json: unknown field "):
			fieldName := strings.Trim(strings.TrimPrefix(err.Error(), "json: unknown field "), `"`)
			msg := fmt.Sprintf("Request body contains unknown field '%s'", fieldName)
			return errors.E(
				errors.WithOp(op), errors.InvalidInput, errors.WithUserMsg(msg), errors.WithErr(err),
			)

		// An io.EOF error is returned by Decode() if the request body is
		// empty.
		case errors.Is(err, io.EOF):
			msg := "Request body must not be empty"
			return errors.E(
				errors.WithOp(op), errors.InvalidInput, errors.WithUserMsg(msg), errors.WithErr(err),
			)

		// Catch the error caused by the request body being too large. Again
		// there is an open issue regarding turning this into a sentinel
		// error at https://github.com/golang/go/issues/30715.
		case err.Error() == "http: request body too large":
			return errors.E(errors.WithOp(op), ErrKindRequestEntityTooLarge, errors.WithErr(err))
		}

		return errors.E(errors.WithOp(op), errors.Internal, errors.WithErr(err))
	}

	// Call decode again, using a pointer to an empty anonymous struct as
	// the destination. If the request body only contained a single JSON
	// object this will return an io.EOF error. So if we get anything else,
	// we know that there is additional data in the request body.
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		msg := "Request body must only contain a single JSON object"
		return errors.E(
			errors.WithOp(op), errors.InvalidInput, errors.WithUserMsg(msg), errors.WithErr(err),
		)
	}

	return nil
}

// checkContentType checks that the Content-Type header is present and
// has the value application/json. The check is skipped if
// SkipCheckContentType is true.
func (j JSONDecoder) checkContentType(r *http.Request) error {
	if j.SkipCheckContentType {
		return nil
	}

	if ct := r.Header.Get("Content-Type"); !isJSONContent(ct) {
		return errors.E(
			ErrKindUnsupportedMediaType,
			errors.WithTextf("Content-Type header '%s' is not application/json", ct),
		)
	}

	return nil
}

// newDecoder sets up the decoder and calls the DisallowUnknownFields(),
// UseNumber() methods on it.
func (j JSONDecoder) newDecoder(r io.Reader) *json.Decoder {
	dec := json.NewDecoder(r)

	if j.UseNumber {
		dec.UseNumber()
	}

	// This will cause Decode() to return a "json: unknown field ..." error if
	// it encounters any extra unexpected fields in the JSON. Strictly speaking,
	// it returns an error for "keys which do not match any non-ignored,
	// exported fields in the destination".
	if j.DisallowUnknownFields {
		dec.DisallowUnknownFields()
	}

	return dec
}

// Source: https://github.com/go-resty/resty/blob/v2.2.0/client.go#L64
var jsonCheck = regexp.MustCompile(`(?i:(application|text)/(json|.*\+json|json\-.*)(;|$))`)

func isJSONContent(ct string) bool { return jsonCheck.MatchString(ct) }
