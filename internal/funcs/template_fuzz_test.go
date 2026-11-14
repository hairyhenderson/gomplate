package funcs

import (
	"context"
	"io"
	"maps"
	"math"
	"math/cmplx"
	"reflect"
	"regexp"
	"slices"
	"strings"
	"sync/atomic"
	"testing"
	"text/template"
	"text/template/parse"

	"github.com/hairyhenderson/gomplate/v5/conv"
	"github.com/stretchr/testify/assert"
)

// fuzzFuncs returns the template functions that don't do any I/O, and the
// names they're called by - top-level aliases, and "namespace.Method", which
// is called as "namespace_Method" so the function can be wrapped. Every
// function sets panicked if it panics, since text/template turns a panic into
// an ordinary error.
func fuzzFuncs(panicked *atomic.Bool) (template.FuncMap, []string) {
	ctx := context.Background()

	// random functions are left out, since every new pattern grows a cache
	// that only the random package can clear - see FuzzStringRE
	all := map[string]any{}
	for _, create := range []func(context.Context) map[string]any{
		CreateBase64Funcs, CreateCollFuncs, CreateConvFuncs, CreateDataFuncs,
		CreateFilePathFuncs, CreateMathFuncs, CreatePathFuncs, CreateReFuncs,
		CreateSemverFuncs, CreateStringFuncs, CreateTestFuncs,
		CreateTimeFuncs, CreateUUIDFuncs,
	} {
		maps.Copy(all, create(ctx))
	}

	// a jq program can loop forever by design, and a CUE program can ask for
	// as much memory as it likes
	delete(all, "jq")
	delete(all, "cue")

	fm := template.FuncMap{}
	names := []string{}
	add := func(name string, fn reflect.Value) {
		fm[strings.ReplaceAll(name, ".", "_")] = recordPanics(fn, panicked).Interface()
		names = append(names, name)
	}

	for k, v := range all {
		ns, ok := v.(func() any)
		if !ok {
			add(k, reflect.ValueOf(v))
			continue
		}

		nv := reflect.ValueOf(ns())
		for i := range nv.NumMethod() {
			m := nv.Type().Method(i).Name
			if k == "coll" && m == "JQ" || k == "data" && m == "CUE" {
				continue
			}
			add(k+"."+m, nv.Method(i))
		}
	}
	slices.Sort(names)

	return fm, names
}

// recordPanics wraps fn so that it sets panicked if it panics.
func recordPanics(fn reflect.Value, panicked *atomic.Bool) reflect.Value {
	return reflect.MakeFunc(fn.Type(), func(args []reflect.Value) []reflect.Value {
		defer func() {
			if r := recover(); r != nil {
				panicked.Store(true)
				panic(r)
			}
		}()

		if fn.Type().IsVariadic() {
			return fn.CallSlice(args)
		}
		return fn.Call(args)
	})
}

// arguments that only hold literals keep each input to one call, so
// fuzzing doesn't just find ways to ask for a huge amount of work
var (
	notLiteral = regexp.MustCompile(`[{}()|$]|\brange\b|\bwith\b|\bdefine\b|\btemplate\b|\bblock\b`)
	number     = regexp.MustCompile(`[0-9.][0-9a-zA-Z.,_]*`)
	exponent   = regexp.MustCompile(`[0-9.][eEpP]`)
	digit      = regexp.MustCompile(`[0-9a-fA-F]`)
)

// bigNumber reports whether s holds a number big enough to ask for a lot of
// work, such as 99999, 0xfffff, 999,999,999, or 1.e9.
func bigNumber(s string) bool {
	for _, n := range number.FindAllString(s, -1) {
		if exponent.MatchString(n) || len(digit.FindAllString(n, -1)) > 4 {
			return true
		}
	}
	return false
}

// bigConv reports whether s converts to a big number, the same way functions
// convert their arguments, so spellings like "0xAp26" can't hide one.
func bigConv(s string) bool {
	if i, err := conv.ToInt64(s); err == nil && (i > 9999 || i < -9999) {
		return true
	}
	if f, err := conv.ToFloat64(s); err == nil && math.Abs(f) > 9999 {
		return true
	}
	return false
}

// bigArgs reports whether any of the parsed template's arguments holds a big
// number. Strings are checked after they're unquoted, so escapes like
// "\x39\x39\x39\x39\x39" can't hide one.
func bigArgs(tmpl *template.Template) bool {
	for _, n := range tmpl.Root.Nodes {
		action, ok := n.(*parse.ActionNode)
		if !ok {
			continue
		}
		for _, cmd := range action.Pipe.Cmds {
			for _, arg := range cmd.Args {
				switch a := arg.(type) {
				case *parse.NumberNode:
					// check the value, since a character literal like
					// '\U0010FFFF' is a number too
					if math.Abs(a.Float64) > 9999 || cmplx.Abs(a.Complex128) > 9999 {
						return true
					}
				case *parse.StringNode:
					if bigNumber(a.Text) || bigConv(a.Text) {
						return true
					}
				}
			}
		}
	}
	return false
}

func FuzzTemplateFuncs(f *testing.F) {
	panicked := &atomic.Bool{}
	fm, names := fuzzFuncs(panicked)

	index := func(name string) uint16 {
		if i := slices.Index(names, name); i >= 0 && i <= math.MaxUint16 {
			return uint16(i)
		}
		f.Fatalf("no function %q", name)
		return 0
	}

	f.Add(index("strings.Indent"), `2 "  " "foo"`)
	f.Add(index("math.Seq"), `-1 5`)
	f.Add(index("coll.Sort"), `"a" nil`)
	f.Add(index("data.ToTOML"), `nil`)
	f.Add(index("conv.ToInt64"), `"1,000"`)
	f.Add(index("strings.WordWrap"), `3 "a b c"`)
	f.Add(index("assert"), ``)
	f.Add(index("coll.Pick"), `nil .m`)
	f.Add(index("coll.Omit"), `nil .m`)
	f.Add(index("coll.JSONPath"), `"" .missing`)
	f.Add(index("math.Rem"), `0 0`)
	f.Add(index("merge"), `. nil`)
	f.Add(index("data.ToYAML"), `0i`)

	f.Fuzz(func(t *testing.T, fn uint16, args string) {
		if len(args) > 200 || notLiteral.MatchString(args) {
			t.Skip()
		}

		name := names[int(fn)%len(names)]
		call := strings.ReplaceAll(name, ".", "_") + " " + args
		// the result is assigned, not printed, since printing a map that holds
		// itself (from set, say) recurses until the stack overflows
		tmpl, err := template.New("fuzz").Funcs(fm).Parse("{{ $_ := " + call + " }}")
		if err != nil {
			return
		}
		if bigArgs(tmpl) {
			t.Skip()
		}

		// a map for the functions that take one
		data := map[string]any{"m": map[string]any{"a": 1, "b": []any{"c"}}}

		panicked.Store(false)
		err = tmpl.Execute(io.Discard, data)
		assert.False(t, panicked.Load(), "%s %s panicked: %v", name, args, err)
	})
}
