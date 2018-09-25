package funcs

import (
	"net/url"
	"sync"
	"text/template"

	"github.com/hairyhenderson/gomplate/conv"
)

var (
	convNS     *ConvFuncs
	convNSInit sync.Once
)

// ConvNS -
func ConvNS() *ConvFuncs {
	convNSInit.Do(func() { convNS = &ConvFuncs{} })
	return convNS
}

// AddConvFuncs -
func AddConvFuncs(f map[string]interface{}) {
	f["conv"] = ConvNS

	f["urlParse"] = ConvNS().URL
	f["bool"] = ConvNS().Bool
	f["has"] = ConvNS().Has
	f["slice"] = ConvNS().Slice
	f["join"] = ConvNS().Join
	f["default"] = ConvNS().Default
	f["dict"] = ConvNS().Dict
}

// ConvFuncs -
type ConvFuncs struct{}

// Bool -
func (f *ConvFuncs) Bool(s interface{}) bool {
	return conv.Bool(conv.ToString(s))
}

// ToBool -
func (f *ConvFuncs) ToBool(in interface{}) bool {
	return conv.ToBool(in)
}

// ToBools -
func (f *ConvFuncs) ToBools(in ...interface{}) []bool {
	return conv.ToBools(in...)
}

// Slice -
func (f *ConvFuncs) Slice(args ...interface{}) []interface{} {
	return conv.Slice(args...)
}

// Join -
func (f *ConvFuncs) Join(in interface{}, sep string) (string, error) {
	return conv.Join(in, sep)
}

// Has -
func (f *ConvFuncs) Has(in interface{}, key string) bool {
	return conv.Has(in, key)
}

// ParseInt -
func (f *ConvFuncs) ParseInt(s interface{}, base, bitSize int) int64 {
	return conv.MustParseInt(conv.ToString(s), base, bitSize)
}

// ParseFloat -
func (f *ConvFuncs) ParseFloat(s interface{}, bitSize int) float64 {
	return conv.MustParseFloat(conv.ToString(s), bitSize)
}

// ParseUint -
func (f *ConvFuncs) ParseUint(s interface{}, base, bitSize int) uint64 {
	return conv.MustParseUint(conv.ToString(s), base, bitSize)
}

// Atoi -
func (f *ConvFuncs) Atoi(s interface{}) int {
	return conv.MustAtoi(conv.ToString(s))
}

// URL -
func (f *ConvFuncs) URL(s interface{}) (*url.URL, error) {
	return url.Parse(conv.ToString(s))
}

// ToInt64 -
func (f *ConvFuncs) ToInt64(in interface{}) int64 {
	return conv.ToInt64(in)
}

// ToInt -
func (f *ConvFuncs) ToInt(in interface{}) int {
	return conv.ToInt(in)
}

// ToInt64s -
func (f *ConvFuncs) ToInt64s(in ...interface{}) []int64 {
	return conv.ToInt64s(in...)
}

// ToInts -
func (f *ConvFuncs) ToInts(in ...interface{}) []int {
	return conv.ToInts(in...)
}

// ToFloat64 -
func (f *ConvFuncs) ToFloat64(in interface{}) float64 {
	return conv.ToFloat64(in)
}

// ToFloat64s -
func (f *ConvFuncs) ToFloat64s(in ...interface{}) []float64 {
	return conv.ToFloat64s(in...)
}

// ToString -
func (f *ConvFuncs) ToString(in interface{}) string {
	return conv.ToString(in)
}

// ToStrings -
func (f *ConvFuncs) ToStrings(in ...interface{}) []string {
	return conv.ToStrings(in...)
}

// Default -
func (f *ConvFuncs) Default(def, in interface{}) interface{} {
	if truth, ok := template.IsTrue(in); truth && ok {
		return in
	}
	return def
}

// Dict -
func (f *ConvFuncs) Dict(in ...interface{}) (map[string]interface{}, error) {
	return conv.Dict(in...)
}
