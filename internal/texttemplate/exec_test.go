// Taken and adapted from the stdlib text/template/exec_test.go.
// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package texttemplate

import (
	"bytes"
	"testing"
	gotemplate "text/template"
)

// T has lots of interesting pieces to use to test execution.
type T struct {
	Tmpl   *gotemplate.Template
	Empty3 any
	S      string
	SI     []int
	SICap  []int
	AI     [3]int
}

var tVal = &T{
	S:      "xyz",
	SI:     []int{3, 4, 5},
	SICap:  make([]int, 5, 10),
	AI:     [3]int{3, 4, 5},
	Empty3: []int{7, 8},
	Tmpl:   gotemplate.Must(gotemplate.New("").Parse("test template")),
}

//nolint:govet
type execTest struct {
	name   string
	input  string
	output string
	data   any
	ok     bool
}

var execTests = []execTest{
	// Slicing.
	{"slice[:]", "{{slice .SI}}", "[3 4 5]", tVal, true},
	{"slice[1:]", "{{slice .SI 1}}", "[4 5]", tVal, true},
	{"slice[1:2]", "{{slice .SI 1 2}}", "[4]", tVal, true},
	{"slice[-1:]", "{{slice .SI -1}}", "", tVal, false},
	{"slice[1:-2]", "{{slice .SI 1 -2}}", "", tVal, false},
	{"slice[1:2:-1]", "{{slice .SI 1 2 -1}}", "", tVal, false},
	{"slice[2:1]", "{{slice .SI 2 1}}", "", tVal, false},
	{"slice[2:2:1]", "{{slice .SI 2 2 1}}", "", tVal, false},
	{"out of range", "{{slice .SI 4 5}}", "", tVal, false},
	{"out of range", "{{slice .SI 2 2 5}}", "", tVal, false},
	{"len(s) < indexes < cap(s)", "{{slice .SICap 6 10}}", "[0 0 0 0]", tVal, true},
	{"len(s) < indexes < cap(s)", "{{slice .SICap 6 10 10}}", "[0 0 0 0]", tVal, true},
	{"indexes > cap(s)", "{{slice .SICap 10 11}}", "", tVal, false},
	{"indexes > cap(s)", "{{slice .SICap 6 10 11}}", "", tVal, false},
	{"array[:]", "{{slice .AI}}", "[3 4 5]", tVal, true},
	{"array[1:]", "{{slice .AI 1}}", "[4 5]", tVal, true},
	{"array[1:2]", "{{slice .AI 1 2}}", "[4]", tVal, true},
	{"string[:]", "{{slice .S}}", "xyz", tVal, true},
	{"string[0:1]", "{{slice .S 0 1}}", "x", tVal, true},
	{"string[1:]", "{{slice .S 1}}", "yz", tVal, true},
	{"string[1:2]", "{{slice .S 1 2}}", "y", tVal, true},
	{"out of range", "{{slice .S 1 5}}", "", tVal, false},
	{"3-index slice of string", "{{slice .S 1 2 2}}", "", tVal, false},
	{"slice of an interface field", "{{slice .Empty3 0 1}}", "[7]", tVal, true},
}

func testExecute(execTests []execTest, template *gotemplate.Template, t *testing.T) {
	b := new(bytes.Buffer)
	funcs := gotemplate.FuncMap{"slice": GoSlice}

	for _, test := range execTests {
		var tmpl *gotemplate.Template
		var err error
		if template == nil {
			tmpl, err = gotemplate.New(test.name).Funcs(funcs).Parse(test.input)
		} else {
			tmpl, err = template.New(test.name).Funcs(funcs).Parse(test.input)
		}
		if err != nil {
			t.Errorf("%s: parse error: %s", test.name, err)
			continue
		}
		b.Reset()
		err = tmpl.Execute(b, test.data)
		switch {
		case !test.ok && err == nil:
			t.Errorf("%s: expected error; got none", test.name)
			continue
		case test.ok && err != nil:
			t.Errorf("%s: unexpected execute error: %s", test.name, err)
			continue
		case !test.ok && err != nil:
			// expected error, got one
		}
		result := b.String()
		if result != test.output {
			t.Errorf("%s: expected\n\t%q\ngot\n\t%q", test.name, test.output, result)
		}
	}
}

func TestExecute(t *testing.T) {
	testExecute(execTests, nil, t)
}
