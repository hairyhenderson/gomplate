package funcs

import (
	"context"
	"errors"
	"math"
	"reflect"
	"regexp"
	"slices"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
)

// fuzzFuncs returns the template functions that don't do any I/O, and the
// names they can be called by - top-level aliases, and "namespace.Method".
func fuzzFuncs() (template.FuncMap, []string) {
	ctx := context.Background()

	fm := template.FuncMap{}
	for _, create := range []func(context.Context) map[string]any{
		CreateBase64Funcs, CreateCollFuncs, CreateConvFuncs, CreateDataFuncs,
		CreateFilePathFuncs, CreateMathFuncs, CreatePathFuncs,
		CreateRandomFuncs, CreateReFuncs, CreateSemverFuncs,
		CreateStringFuncs, CreateTestFuncs, CreateTimeFuncs, CreateUUIDFuncs,
	} {
		for k, v := range create(ctx) {
			fm[k] = v
		}
	}

	// a jq program can loop forever by design
	delete(fm, "jq")

	names := []string{}
	for k, v := range fm {
		ns, ok := v.(func() any)
		if !ok {
			names = append(names, k)
			continue
		}

		t := reflect.TypeOf(ns())
		for i := range t.NumMethod() {
			m := t.Method(i).Name
			if k == "coll" && m == "JQ" {
				continue
			}
			names = append(names, k+"."+m)
		}
	}
	slices.Sort(names)

	return fm, names
}

// arguments that only hold literals keep each input to one call, so
// fuzzing doesn't just find ways to ask for a huge amount of work
var (
	notLiteral = regexp.MustCompile(`[{}()|$]|\brange\b|\bwith\b|\bdefine\b|\btemplate\b|\bblock\b`)
	bigNumber  = regexp.MustCompile(`[0-9]{4,}|[0-9][eEpPxXoObB_]`)
	panicked   = regexp.MustCompile(`runtime error|reflect: |interface conversion`)
)

// limitWriter fails writes past n bytes.
type limitWriter struct{ n int }

func (w *limitWriter) Write(p []byte) (int, error) {
	w.n -= len(p)
	if w.n < 0 {
		return 0, errors.New("output too long")
	}
	return len(p), nil
}

func FuzzTemplateFuncs(f *testing.F) {
	fm, names := fuzzFuncs()

	index := func(name string) uint16 {
		if i := slices.Index(names, name); i >= 0 && i <= math.MaxUint16 {
			return uint16(i)
		}
		f.Fatalf("no function %q", name)
		return 0
	}

	f.Add(index("strings.Indent"), `2 "  " "foo"`)
	f.Add(index("math.Seq"), `-1 5`)
	f.Add(index("random.String"), `-1`)
	f.Add(index("coll.Sort"), `"a" nil`)
	f.Add(index("data.ToTOML"), `nil`)
	f.Add(index("conv.ToInt64"), `"1,000"`)
	f.Add(index("strings.WordWrap"), `3 "a b c"`)
	f.Add(index("assert"), ``)
	f.Add(index("coll.Pick"), `nil .m`)
	f.Add(index("coll.Omit"), `nil .m`)
	f.Add(index("coll.JSONPath"), `"" .missing`)
	f.Add(index("math.Rem"), `0 0`)

	f.Fuzz(func(t *testing.T, fn uint16, args string) {
		if len(args) > 200 || notLiteral.MatchString(args) || bigNumber.MatchString(args) {
			t.Skip()
		}

		name := names[int(fn)%len(names)]
		tmpl, err := template.New("fuzz").Funcs(fm).Parse("{{ " + name + " " + args + " }}")
		if err != nil {
			return
		}

		// a map for the functions that take one
		data := map[string]any{"m": map[string]any{"a": 1, "b": []any{"c"}}}

		err = tmpl.Execute(&limitWriter{n: 1 << 20}, data)
		if err != nil && panicked.MatchString(err.Error()) {
			assert.Fail(t, "panic", "%s %s: %v", name, args, err)
		}
	})
}
