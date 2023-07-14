package funcs

// Namespace strings contains mostly wrappers of equivalently-named
// functions in the standard library `strings` package, with
// differences in argument order where it makes pipelining
// in templates easier.

import (
	"context"
	"fmt"
	"reflect"
	"unicode/utf8"

	"github.com/Masterminds/goutils"
	"github.com/Masterminds/semver/v3"
	"github.com/flanksource/gomplate/v3/conv"
	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"strings"

	gompstrings "github.com/flanksource/gomplate/v3/strings"
	"github.com/gosimple/slug"
)

// StrNS -
// Deprecated: don't use
func StrNS() *StringFuncs {
	return &StringFuncs{}
}

// AddStringFuncs -
// Deprecated: use CreateStringFuncs instead
func AddStringFuncs(f map[string]interface{}) {
	for k, v := range CreateStringFuncs(context.Background()) {
		f[k] = v
	}
}

// CreateStringFuncs -
func CreateStringFuncs(ctx context.Context) map[string]interface{} {
	f := map[string]interface{}{}

	ns := &StringFuncs{ctx, language.Und}
	f["strings"] = func() interface{} { return ns }

	f["replaceAll"] = ns.ReplaceAll
	f["title"] = ns.Title
	f["toUpper"] = ns.ToUpper
	f["toLower"] = ns.ToLower
	f["trim"] = ns.TrimSpace
	f["indent"] = ns.Indent
	f["quote"] = ns.Quote
	f["shellQuote"] = ns.ShellQuote
	f["squote"] = ns.Squote

	// these are legacy aliases with non-pipelinable arg order
	f["contains"] = strings.Contains
	f["hasPrefix"] = strings.HasPrefix
	f["hasSuffix"] = strings.HasSuffix
	f["startsWith"] = strings.HasPrefix
	f["endsWith"] = strings.HasSuffix
	f["split"] = strings.Split
	f["splitN"] = strings.SplitN
	f["trim"] = strings.Trim
	f["humanDuration"] = gompstrings.HumanDuration
	f["humanSize"] = gompstrings.HumanBytes
	f["semver"] = gompstrings.Semver
	f["semverCompare"] = gompstrings.SemverCompare

	return f
}

func (ns StringFuncs) HumanDuration(in interface{}) (string, error) {
	return gompstrings.HumanDuration(in), nil
}

func (ns StringFuncs) HumanSize(in interface{}) (string, error) {
	return gompstrings.HumanBytes(in), nil
}

func (ns StringFuncs) Semver(in string) (*semver.Version, error) {
	return gompstrings.Semver(in)
}

func (ns StringFuncs) SemverMap(in string) (map[string]string, error) {
	v, err := gompstrings.Semver(in)
	if err != nil {
		return nil, err
	}

	res := map[string]string{
		"major":      fmt.Sprint(v.Major()),
		"minor":      fmt.Sprint(v.Minor()),
		"patch":      fmt.Sprint(v.Patch()),
		"prerelease": v.Prerelease(),
		"metadata":   v.Metadata(),
		"original":   v.Original(),
	}

	return res, nil
}

func (ns StringFuncs) SemverCompare(v1, v2 string) (bool, error) {
	return gompstrings.SemverCompare(v1, v2)
}

// StringFuncs -
type StringFuncs struct {
	ctx context.Context

	// tag - the selected BCP 47 language tag. Currently gomplate only supports
	// Und (undetermined)
	tag language.Tag
}

// Abbrev -
func (StringFuncs) Abbrev(args ...interface{}) (string, error) {
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
func (StringFuncs) ReplaceAll(old, new string, s interface{}) string {
	return strings.ReplaceAll(conv.ToString(s), old, new)
}

// Contains -
func (StringFuncs) Contains(substr string, s interface{}) bool {
	return strings.Contains(conv.ToString(s), substr)
}

// HasPrefix -
func (StringFuncs) HasPrefix(prefix string, s interface{}) bool {
	return strings.HasPrefix(conv.ToString(s), prefix)
}

// HasSuffix -
func (StringFuncs) HasSuffix(suffix string, s interface{}) bool {
	return strings.HasSuffix(conv.ToString(s), suffix)
}

// Repeat -
func (StringFuncs) Repeat(count int, s interface{}) (string, error) {
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
//
// Deprecated: use coll.Sort instead
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
func (StringFuncs) Split(sep string, s interface{}) []string {
	return strings.Split(conv.ToString(s), sep)
}

// SplitN -
func (StringFuncs) SplitN(sep string, n int, s interface{}) []string {
	return strings.SplitN(conv.ToString(s), sep, n)
}

// Trim -
func (StringFuncs) Trim(cutset string, s interface{}) string {
	return strings.Trim(conv.ToString(s), cutset)
}

// TrimPrefix -
func (StringFuncs) TrimPrefix(cutset string, s interface{}) string {
	return strings.TrimPrefix(conv.ToString(s), cutset)
}

// TrimSuffix -
func (StringFuncs) TrimSuffix(cutset string, s interface{}) string {
	return strings.TrimSuffix(conv.ToString(s), cutset)
}

// Title -
func (f *StringFuncs) Title(s interface{}) string {
	return cases.Title(f.tag).String(conv.ToString(s))
}

// ToUpper -
func (f *StringFuncs) ToUpper(s interface{}) string {
	return cases.Upper(f.tag).String(conv.ToString(s))
}

// ToLower -
func (f *StringFuncs) ToLower(s interface{}) string {
	return cases.Lower(f.tag).String(conv.ToString(s))
}

// TrimSpace -
func (StringFuncs) TrimSpace(s interface{}) string {
	return strings.TrimSpace(conv.ToString(s))
}

// Trunc -
func (StringFuncs) Trunc(length int, s interface{}) string {
	return gompstrings.Trunc(length, conv.ToString(s))
}

// Indent -
func (StringFuncs) Indent(args ...interface{}) (string, error) {
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
				return "", errors.New("indent: invalid arguments")
			}
			indent = " "
		}
	case 3:
		width, ok = args[0].(int)
		if !ok {
			return "", errors.New("indent: invalid arguments")
		}
		indent, ok = args[1].(string)
		if !ok {
			return "", errors.New("indent: invalid arguments")
		}
	}
	return gompstrings.Indent(width, indent, input), nil
}

// Slug -
func (StringFuncs) Slug(in interface{}) string {
	return slug.Make(conv.ToString(in))
}

// Quote -
func (StringFuncs) Quote(in interface{}) string {
	return fmt.Sprintf("%q", conv.ToString(in))
}

// ShellQuote -
func (StringFuncs) ShellQuote(in interface{}) string {
	val := reflect.ValueOf(in)
	switch val.Kind() {
	case reflect.Array, reflect.Slice:
		var sb strings.Builder
		max := val.Len()
		for n := 0; n < max; n++ {
			sb.WriteString(gompstrings.ShellQuote(conv.ToString(val.Index(n))))
			if n+1 != max {
				sb.WriteRune(' ')
			}
		}
		return sb.String()
	}
	return gompstrings.ShellQuote(conv.ToString(in))
}

// Squote -
func (StringFuncs) Squote(in interface{}) string {
	s := conv.ToString(in)
	s = strings.ReplaceAll(s, `'`, `''`)
	return fmt.Sprintf("'%s'", s)
}

// SnakeCase -
func (StringFuncs) SnakeCase(in interface{}) (string, error) {
	return gompstrings.SnakeCase(conv.ToString(in)), nil
}

// CamelCase -
func (StringFuncs) CamelCase(in interface{}) (string, error) {
	return gompstrings.CamelCase(conv.ToString(in)), nil
}

// KebabCase -
func (StringFuncs) KebabCase(in interface{}) (string, error) {
	return gompstrings.KebabCase(conv.ToString(in)), nil
}

// WordWrap -
func (StringFuncs) WordWrap(args ...interface{}) (string, error) {
	if len(args) == 0 || len(args) > 3 {
		return "", errors.Errorf("expected 1, 2, or 3 args, got %d", len(args))
	}
	in := conv.ToString(args[len(args)-1])

	opts := gompstrings.WordWrapOpts{}
	if len(args) == 2 {
		switch a := (args[0]).(type) {
		case string:
			opts.LBSeq = a
		default:
			opts.Width = uint(conv.ToInt(a))
		}
	}
	if len(args) == 3 {
		opts.Width = uint(conv.ToInt(args[0]))
		opts.LBSeq = conv.ToString(args[1])
	}
	return gompstrings.WordWrap(in, opts), nil
}

// RuneCount - like len(s), but for runes
func (StringFuncs) RuneCount(args ...interface{}) (int, error) {
	s := ""
	for _, arg := range args {
		s += conv.ToString(arg)
	}
	return utf8.RuneCountInString(s), nil
}
