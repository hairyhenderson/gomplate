package funcs

// Namespace strings contains mostly wrappers of equivalently-named
// functions in the standard library `strings` package, with
// differences in argument order where it makes pipelining
// in templates easier.

import (
	"log"
	"sync"

	"strings"

	gompstrings "github.com/hairyhenderson/gomplate/strings"
)

var (
	strNS     *StringFuncs
	strNSInit sync.Once
)

// StrNS -
func StrNS() *StringFuncs {
	strNSInit.Do(func() { strNS = &StringFuncs{} })
	return strNS
}

// AddStringFuncs -
func AddStringFuncs(f map[string]interface{}) {
	f["strings"] = StrNS

	f["replaceAll"] = StrNS().ReplaceAll
	f["title"] = StrNS().Title
	f["toUpper"] = StrNS().ToUpper
	f["toLower"] = StrNS().ToLower
	f["trimSpace"] = StrNS().TrimSpace
	f["indent"] = StrNS().Indent

	// these are legacy aliases with non-pipelinable arg order
	f["contains"] = strings.Contains
	f["hasPrefix"] = strings.HasPrefix
	f["hasSuffix"] = strings.HasSuffix
	f["split"] = strings.Split
	f["splitN"] = strings.SplitN
	f["trim"] = strings.Trim
}

// StringFuncs -
type StringFuncs struct{}

// ReplaceAll -
func (f *StringFuncs) ReplaceAll(old, new, s string) string {
	return strings.Replace(s, old, new, -1)
}

// Contains -
func (f *StringFuncs) Contains(substr, s string) bool {
	return strings.Contains(s, substr)
}

// HasPrefix -
func (f *StringFuncs) HasPrefix(prefix, s string) bool {
	return strings.HasPrefix(s, prefix)
}

// HasSuffix -
func (f *StringFuncs) HasSuffix(suffix, s string) bool {
	return strings.HasSuffix(s, suffix)
}

// Split -
func (f *StringFuncs) Split(sep, s string) []string {
	return strings.Split(s, sep)
}

// SplitN -
func (f *StringFuncs) SplitN(sep string, n int, s string) []string {
	return strings.SplitN(s, sep, n)
}

// Trim -
func (f *StringFuncs) Trim(cutset, s string) string {
	return strings.Trim(s, cutset)
}

// Trim Prefix-
func (f *StringFuncs) TrimPrefix(cutset, s string) string {
	return strings.TrimPrefix(s, cutset)
}

// Title -
func (f *StringFuncs) Title(s string) string {
	return strings.Title(s)
}

// ToUpper -
func (f *StringFuncs) ToUpper(s string) string {
	return strings.ToUpper(s)
}

// ToLower -
func (f *StringFuncs) ToLower(s string) string {
	return strings.ToLower(s)
}

// TrimSpace -
func (f *StringFuncs) TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

// Indent -
func (f *StringFuncs) Indent(args ...interface{}) string {
	input, ok := args[len(args)-1].(string)
	if !ok {
		log.Fatal("Indent: invalid arguments")
	}
	indent := " "
	width := 1
	switch len(args) {
	case 2:
		indent, ok = args[0].(string)
		if !ok {
			width, ok = args[0].(int)
			if !ok {
				log.Fatal("Indent: invalid arguments")
			}
			indent = " "
		}
	case 3:
		width, ok = args[0].(int)
		if !ok {
			log.Fatal("Indent: invalid arguments")
		}
		indent, ok = args[1].(string)
		if !ok {
			log.Fatal("Indent: invalid arguments")
		}
	}
	return gompstrings.Indent(width, indent, input)
}
