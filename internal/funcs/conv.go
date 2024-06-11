package funcs

import (
	"context"
	"net/url"
	"strconv"
	"text/template"

	"github.com/hairyhenderson/gomplate/v4/conv"
	"github.com/hairyhenderson/gomplate/v4/internal/deprecated"
)

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

// Join -
func (ConvFuncs) Join(in interface{}, sep string) (string, error) {
	return conv.Join(in, sep)
}

// ParseInt -
func (ConvFuncs) ParseInt(s interface{}, base, bitSize int) (int64, error) {
	return strconv.ParseInt(conv.ToString(s), base, bitSize)
}

// ParseFloat -
func (ConvFuncs) ParseFloat(s interface{}, bitSize int) (float64, error) {
	return strconv.ParseFloat(conv.ToString(s), bitSize)
}

// ParseUint -
func (ConvFuncs) ParseUint(s interface{}, base, bitSize int) (uint64, error) {
	return strconv.ParseUint(conv.ToString(s), base, bitSize)
}

// Atoi -
func (ConvFuncs) Atoi(s interface{}) (int, error) {
	return strconv.Atoi(conv.ToString(s))
}

// URL -
func (ConvFuncs) URL(s interface{}) (*url.URL, error) {
	return url.Parse(conv.ToString(s))
}

// ToInt64 -
func (ConvFuncs) ToInt64(in interface{}) (int64, error) {
	return conv.ToInt64(in)
}

// ToInt -
func (ConvFuncs) ToInt(in interface{}) (int, error) {
	return conv.ToInt(in)
}

// ToInt64s -
func (ConvFuncs) ToInt64s(in ...interface{}) ([]int64, error) {
	return conv.ToInt64s(in...)
}

// ToInts -
func (ConvFuncs) ToInts(in ...interface{}) ([]int, error) {
	return conv.ToInts(in...)
}

// ToFloat64 -
func (ConvFuncs) ToFloat64(in interface{}) (float64, error) {
	return conv.ToFloat64(in)
}

// ToFloat64s -
func (ConvFuncs) ToFloat64s(in ...interface{}) ([]float64, error) {
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
