package funcs

import (
	"net/url"
	"sync"

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
}

// ConvFuncs -
type ConvFuncs struct{}

// Bool -
func (f *ConvFuncs) Bool(s string) bool {
	return conv.Bool(s)
}

// Slice -
func (f *ConvFuncs) Slice(args ...interface{}) []interface{} {
	return conv.Slice(args...)
}

// Join -
func (f *ConvFuncs) Join(in interface{}, sep string) string {
	return conv.Join(in, sep)
}

// Has -
func (f *ConvFuncs) Has(in interface{}, key string) bool {
	return conv.Has(in, key)
}

// ParseInt -
func (f *ConvFuncs) ParseInt(s string, base, bitSize int) int64 {
	return conv.MustParseInt(s, base, bitSize)
}

// ParseFloat -
func (f *ConvFuncs) ParseFloat(s string, bitSize int) float64 {
	return conv.MustParseFloat(s, bitSize)
}

// ParseUint -
func (f *ConvFuncs) ParseUint(s string, base, bitSize int) uint64 {
	return conv.MustParseUint(s, base, bitSize)
}

// Atoi -
func (f *ConvFuncs) Atoi(s string) int {
	return conv.MustAtoi(s)
}

// URL -
func (f *ConvFuncs) URL(s string) (*url.URL, error) {
	return url.Parse(s)
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
	return conv.ToInt64s(in)
}

// ToInts -
func (f *ConvFuncs) ToInts(in ...interface{}) []int {
	return conv.ToInts(in)
}

// ToFloat64 -
func (f *ConvFuncs) ToFloat64(in interface{}) float64 {
	return conv.ToFloat64(in)
}

// ToFloat64s -
func (f *ConvFuncs) ToFloat64s(in ...interface{}) []float64 {
	return conv.ToFloat64s(in)
}

// ToString -
func (f *ConvFuncs) ToString(in interface{}) string {
	return conv.ToString(in)
}
