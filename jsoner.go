package xgo

// JSONer is implemented by any value that has a JSON method.
type JSONer interface {
	JSON() interface{}
}
