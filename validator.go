package xgo

// Validator is implemented by any value which has a Validate method.
type Validator interface {
	// Validate validates the given value.
	Validate(interface{}) error
}

// ValidatorFunc type is an adapter to allow the use of ordinary
// functions as a Validator. If f is a function with the appropriate
// signature, ValidatorFunc(f) is an Validator that calls f.
type ValidatorFunc func(interface{}) error

// Validate calls f(v).
func (f ValidatorFunc) Validate(v interface{}) error {
	return f(v)
}
