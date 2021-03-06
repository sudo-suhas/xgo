package httputil

import (
	"net/http"

	"github.com/sudo-suhas/xgo"
	"github.com/sudo-suhas/xgo/errors"
)

// Decoder is implemented by any value which has a Decode method.
type Decoder interface {
	// Decode decodes the HTTP request into the given value.
	Decode(r *http.Request, v interface{}) error
}

// DecodeFunc type is an adapter to allow the use of ordinary functions
// as a Decoder. If f is a function with the appropriate
// signature, DecodeFunc(f) is a Decoder that calls f.
type DecodeFunc func(r *http.Request, v interface{}) error

// Decode calls f(r, v).
func (f DecodeFunc) Decode(r *http.Request, v interface{}) error {
	return f(r, v)
}

// DecoderMiddleware describes a middleware function for Decoder.
type DecoderMiddleware func(Decoder) Decoder

// ValidatingDecoderMiddleware returns a DecoderMiddleware which
// validates the decoded value.
//
// 	var vd xgo.Validator = MyValidator{}
// 	var dec httputil.Decoder
// 	{
// 		dec = httputil.JSONDecoder{}
// 		dec = httputil.ValidatingDecoderMiddleware(vd)(dec)
// 	}
func ValidatingDecoderMiddleware(vd xgo.Validator) DecoderMiddleware {
	return func(d Decoder) Decoder {
		return DecodeFunc(func(r *http.Request, v interface{}) error {
			const op = "ValidatingDecoderMiddleware"

			if err := d.Decode(r, v); err != nil {
				return err
			}

			if err := vd.Validate(v); err != nil {
				return errors.E(errors.WithOp(op), errors.WithErr(err))
			}

			return nil
		})
	}
}
