package funcs

// Namespace strings contains mostly wrappers of equivalently-named
// functions in the standard library `strings` package, with
// differences in argument order where it makes pipelining
// in templates easier.

import (
	"sync"

	"github.com/Masterminds/goutils"
	"github.com/hairyhenderson/gomplate/conv"
	"github.com/pkg/errors"

	"strings"

	"github.com/gosimple/slug"
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
	f["sort"] = StrNS().Sort

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

// Abbrev -
func (f *StringFuncs) Abbrev(args ...interface{}) (string, error) {
	str := ""
	offset := 0
	maxWidth := 0
	if len(args) < 2 {
		return "", errors.Errorf("abbrev requires a 'maxWidth' and 'input' argument")
	}
	if len(args) == 2 {
		maxWidth = conv.ToInt(args[0])
		str = conv.ToString(args[1])
	}
	if len(args) == 3 {
		offset = conv.ToInt(args[0])
		maxWidth = conv.ToInt(args[1])
		str = conv.ToString(args[2])
	}
	if len(str) <= maxWidth {
		return str, nil
	}
	return goutils.AbbreviateFull(str, offset, maxWidth)
}

// ReplaceAll -
func (f *StringFuncs) ReplaceAll(old, new string, s interface{}) string {
	return strings.Replace(conv.ToString(s), old, new, -1)
}

// Contains -
func (f *StringFuncs) Contains(substr string, s interface{}) bool {
	return strings.Contains(conv.ToString(s), substr)
}

// HasPrefix -
func (f *StringFuncs) HasPrefix(prefix string, s interface{}) bool {
	return strings.HasPrefix(conv.ToString(s), prefix)
}

// HasSuffix -
func (f *StringFuncs) HasSuffix(suffix string, s interface{}) bool {
	return strings.HasSuffix(conv.ToString(s), suffix)
}

// Repeat -
func (f *StringFuncs) Repeat(count int, s interface{}) (string, error) {
	if count < 0 {
		return "", errors.Errorf("negative count %d", count)
	}
	str := conv.ToString(s)
	if count > 0 && len(str)*count/count != len(str) {
		return "", errors.Errorf("count %d too long: causes overflow", count)
	}
	return strings.Repeat(str, count), nil
}

// Sort -
func (f *StringFuncs) Sort(list interface{}) ([]string, error) {
	switch v := list.(type) {
	case []string:
		return gompstrings.Sort(v), nil
	case []interface{}:
		l := len(v)
		b := make([]string, len(v))
		for i := 0; i < l; i++ {
			b[i] = conv.ToString(v[i])
		}
		return gompstrings.Sort(b), nil
	default:
		return nil, errors.Errorf("wrong type for value; expected []string; got %T", list)
	}
}

// Split -
func (f *StringFuncs) Split(sep string, s interface{}) []string {
	return strings.Split(conv.ToString(s), sep)
}

// SplitN -
func (f *StringFuncs) SplitN(sep string, n int, s interface{}) []string {
	return strings.SplitN(conv.ToString(s), sep, n)
}

// Trim -
func (f *StringFuncs) Trim(cutset string, s interface{}) string {
	return strings.Trim(conv.ToString(s), cutset)
}

// TrimPrefix -
func (f *StringFuncs) TrimPrefix(cutset string, s interface{}) string {
	return strings.TrimPrefix(conv.ToString(s), cutset)
}

// TrimSuffix -
func (f *StringFuncs) TrimSuffix(cutset string, s interface{}) string {
	return strings.TrimSuffix(conv.ToString(s), cutset)
}

// Title -
func (f *StringFuncs) Title(s interface{}) string {
	return strings.Title(conv.ToString(s))
}

// ToUpper -
func (f *StringFuncs) ToUpper(s interface{}) string {
	return strings.ToUpper(conv.ToString(s))
}

// ToLower -
func (f *StringFuncs) ToLower(s interface{}) string {
	return strings.ToLower(conv.ToString(s))
}

// TrimSpace -
func (f *StringFuncs) TrimSpace(s interface{}) string {
	return strings.TrimSpace(conv.ToString(s))
}

// Trunc -
func (f *StringFuncs) Trunc(length int, s interface{}) string {
	return gompstrings.Trunc(length, conv.ToString(s))
}

// Indent -
func (f *StringFuncs) Indent(args ...interface{}) (string, error) {
	input := conv.ToString(args[len(args)-1])
	indent := " "
	width := 1
	var ok bool
	switch len(args) {
	case 2:
		indent, ok = args[0].(string)
		if !ok {
			width, ok = args[0].(int)
			if !ok {
				return "", errors.New("Indent: invalid arguments")
			}
			indent = " "
		}
	case 3:
		width, ok = args[0].(int)
		if !ok {
			return "", errors.New("Indent: invalid arguments")
		}
		indent, ok = args[1].(string)
		if !ok {
			return "", errors.New("Indent: invalid arguments")
		}
	}
	return gompstrings.Indent(width, indent, input), nil
}

// Slug -
func (f *StringFuncs) Slug(in interface{}) string {
	return slug.Make(conv.ToString(in))
}
