package funcs

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/hairyhenderson/gomplate/base64"
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
func (f *Base64Funcs) Encode(in interface{}) string {
	b := toBytes(in)
	return base64.Encode(b)
}

// Decode -
func (f *Base64Funcs) Decode(in interface{}) string {
	return string(base64.Decode(toString(in)))
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
	return []byte(fmt.Sprintf("%s", in))
}

func toString(in interface{}) string {
	if s, ok := in.(string); ok {
		return s
	}
	if s, ok := in.(fmt.Stringer); ok {
		return s.String()
	}
	if i, ok := in.(int); ok {
		return strconv.Itoa(i)
	}
	if u, ok := in.(uint64); ok {
		return strconv.FormatUint(u, 10)
	}
	if f, ok := in.(float64); ok {
		return strconv.FormatFloat(f, 'f', -1, 64)
	}
	if b, ok := in.(bool); ok {
		return strconv.FormatBool(b)
	}
	if in == nil {
		return "nil"
	}
	return fmt.Sprintf("%s", in)
}
