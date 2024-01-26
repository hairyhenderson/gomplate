package funcs

import (
	"context"

	"github.com/hairyhenderson/gomplate/v4/base64"
	"github.com/hairyhenderson/gomplate/v4/conv"
)

// CreateBase64Funcs -
func CreateBase64Funcs(ctx context.Context) map[string]interface{} {
	f := map[string]interface{}{}

	ns := &Base64Funcs{ctx}
	f["base64"] = func() interface{} { return ns }

	return f
}

// Base64Funcs -
type Base64Funcs struct {
	ctx context.Context
}

// Encode -
func (Base64Funcs) Encode(in interface{}) (string, error) {
	b := toBytes(in)
	return base64.Encode(b)
}

// Decode -
func (Base64Funcs) Decode(in interface{}) (string, error) {
	out, err := base64.Decode(conv.ToString(in))
	return string(out), err
}

// DecodeBytes -
func (Base64Funcs) DecodeBytes(in interface{}) ([]byte, error) {
	out, err := base64.Decode(conv.ToString(in))
	return out, err
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
