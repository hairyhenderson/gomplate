package funcs

import (
	"sync"

	"github.com/hairyhenderson/gomplate/base64"
	"github.com/hairyhenderson/gomplate/conv"
)

var (
	bf     *Base64Funcs
	bfInit sync.Once
)

// Base64NS - the base64 namespace
func Base64NS() *Base64Funcs {
	bfInit.Do(func() { bf = &Base64Funcs{} })
	return bf
}

// AddBase64Funcs -
func AddBase64Funcs(f map[string]interface{}) {
	f["base64"] = Base64NS
}

// Base64Funcs -
type Base64Funcs struct{}

// Encode -
func (f *Base64Funcs) Encode(in interface{}) (string, error) {
	b := toBytes(in)
	return base64.Encode(b)
}

// Decode -
func (f *Base64Funcs) Decode(in interface{}) (string, error) {
	out, err := base64.Decode(conv.ToString(in))
	return string(out), err
}

type byter interface {
	Bytes() []byte
}

func toBytes(in interface{}) []byte {
	if in == nil {
		return []byte{}
	}
	if s, ok := in.([]byte); ok {
		return s
	}
	if s, ok := in.(byter); ok {
		return s.Bytes()
	}
	if s, ok := in.(string); ok {
		return []byte(s)
	}
	return []byte(conv.ToString(in))
}
