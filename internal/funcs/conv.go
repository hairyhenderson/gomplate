package funcs

import (
	"context"
	"net/url"
	"text/template"

	"github.com/hairyhenderson/gomplate/v4/coll"
	"github.com/hairyhenderson/gomplate/v4/conv"
	"github.com/hairyhenderson/gomplate/v4/internal/deprecated"
)

// ConvNS -
//
// Deprecated: don't use
func ConvNS() *ConvFuncs {
	return &ConvFuncs{}
}

// AddConvFuncs -
//
// Deprecated: use [CreateConvFuncs] instead
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
//
// Deprecated: use [ToBool] instead
func (f *ConvFuncs) Bool(s interface{}) bool {
	deprecated.WarnDeprecated(f.ctx, "conv.Bool is deprecated - use conv.ToBool instead")
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
//
// Deprecated: use [CollFuncs.Slice] instead
func (f *ConvFuncs) Slice(args ...interface{}) []interface{} {
	deprecated.WarnDeprecated(f.ctx, "conv.Slice is deprecated - use coll.Slice instead")
	return coll.Slice(args...)
}

// Join -
func (ConvFuncs) Join(in interface{}, sep string) (string, error) {
	return conv.Join(in, sep)
}

// Has -
//
// Deprecated: use [CollFuncs.Has] instead
func (f *ConvFuncs) Has(in interface{}, key string) bool {
	deprecated.WarnDeprecated(f.ctx, "conv.Has is deprecated - use coll.Has instead")
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
//
// Deprecated: use [CollFuncs.Dict] instead
func (f *ConvFuncs) Dict(in ...interface{}) (map[string]interface{}, error) {
	deprecated.WarnDeprecated(f.ctx, "conv.Dict is deprecated - use coll.Dict instead")
	return coll.Dict(in...)
}
