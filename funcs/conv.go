package funcs

import (
	"context"
	"net/url"
	"text/template"

	"github.com/hairyhenderson/gomplate/v3/coll"
	"github.com/hairyhenderson/gomplate/v3/conv"
)

// ConvNS -
// Deprecated: don't use
func ConvNS() *ConvFuncs {
	return &ConvFuncs{}
}

// AddConvFuncs -
// Deprecated: use CreateConvFuncs instead
func AddConvFuncs(f map[string]interface{}) {
	for k, v := range CreateConvFuncs(context.Background()) {
		f[k] = v
	}
}

// CreateConvFuncs -
func CreateConvFuncs(ctx context.Context) map[string]interface{} {
	ns := &ConvFuncs{ctx}

	f := map[string]interface{}{}
	f["conv"] = func() interface{} { return ns }

	f["urlParse"] = ns.URL
	f["bool"] = ns.Bool
	f["join"] = ns.Join
	f["default"] = ns.Default
	return f
}

// ConvFuncs -
type ConvFuncs struct {
	ctx context.Context
}

// Bool -
func (ConvFuncs) Bool(s interface{}) bool {
	return conv.Bool(conv.ToString(s))
}

// ToBool -
func (ConvFuncs) ToBool(in interface{}) bool {
	return conv.ToBool(in)
}

// ToBools -
func (ConvFuncs) ToBools(in ...interface{}) []bool {
	return conv.ToBools(in...)
}

// Slice -
func (ConvFuncs) Slice(args ...interface{}) []interface{} {
	return coll.Slice(args...)
}

// Join -
func (ConvFuncs) Join(in interface{}, sep string) (string, error) {
	return conv.Join(in, sep)
}

// Has -
func (ConvFuncs) Has(in interface{}, key string) bool {
	return coll.Has(in, key)
}

// ParseInt -
func (ConvFuncs) ParseInt(s interface{}, base, bitSize int) int64 {
	return conv.MustParseInt(conv.ToString(s), base, bitSize)
}

// ParseFloat -
func (ConvFuncs) ParseFloat(s interface{}, bitSize int) float64 {
	return conv.MustParseFloat(conv.ToString(s), bitSize)
}

// ParseUint -
func (ConvFuncs) ParseUint(s interface{}, base, bitSize int) uint64 {
	return conv.MustParseUint(conv.ToString(s), base, bitSize)
}

// Atoi -
func (ConvFuncs) Atoi(s interface{}) int {
	return conv.MustAtoi(conv.ToString(s))
}

// URL -
func (ConvFuncs) URL(s interface{}) (*url.URL, error) {
	return url.Parse(conv.ToString(s))
}

// ToInt64 -
func (ConvFuncs) ToInt64(in interface{}) int64 {
	return conv.ToInt64(in)
}

// ToInt -
func (ConvFuncs) ToInt(in interface{}) int {
	return conv.ToInt(in)
}

// ToInt64s -
func (ConvFuncs) ToInt64s(in ...interface{}) []int64 {
	return conv.ToInt64s(in...)
}

// ToInts -
func (ConvFuncs) ToInts(in ...interface{}) []int {
	return conv.ToInts(in...)
}

// ToFloat64 -
func (ConvFuncs) ToFloat64(in interface{}) float64 {
	return conv.ToFloat64(in)
}

// ToFloat64s -
func (ConvFuncs) ToFloat64s(in ...interface{}) []float64 {
	return conv.ToFloat64s(in...)
}

// ToString -
func (ConvFuncs) ToString(in interface{}) string {
	return conv.ToString(in)
}

// ToStrings -
func (ConvFuncs) ToStrings(in ...interface{}) []string {
	return conv.ToStrings(in...)
}

// Default -
func (ConvFuncs) Default(def, in interface{}) interface{} {
	if truth, ok := template.IsTrue(in); truth && ok {
		return in
	}
	return def
}

// Dict -
func (ConvFuncs) Dict(in ...interface{}) (map[string]interface{}, error) {
	return coll.Dict(in...)
}
