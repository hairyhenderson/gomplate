// Package tmpl contains functions for defining or executing in-line templates.
package tmpl

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"
)

// Template -
type Template struct {
	root       *template.Template
	defaultCtx interface{}
	path       string
}

// New -
func New(root *template.Template, tctx interface{}, path string) *Template {
	return &Template{root, tctx, path}
}

// Path - returns the path to the current template if it came from a file.
// An empty string is returned for inline templates.
func (t *Template) Path() (string, error) {
	return t.path, nil
}

// PathDir - returns the directory of the template, if it came from a file. An empty
// string is returned for inline templates. If the template was loaded from the
// current working directory, "." is returned.
func (t *Template) PathDir() (string, error) {
	if t.path == "" {
		return "", nil
	}
	return filepath.Dir(t.path), nil
}

// Inline - a template function to do inline template processing
//
// Can be called 4 ways:
// {{ tmpl.Inline "inline template" }} - unnamed (single-use) template with default context
// {{ tmpl.Inline "name" "inline template" }} - named template with default context
// {{ tmpl.Inline "inline template" $foo }} - unnamed (single-use) template with given context
// {{ tmpl.Inline "name" "inline template" $foo }} - named template with given context
func (t *Template) Inline(args ...interface{}) (string, error) {
	name, in, ctx, err := t.parseArgs(args...)
	if err != nil {
		return "", err
	}
	return t.inline(name, in, ctx)
}

func (t *Template) inline(name, in string, ctx interface{}) (string, error) {
	tmpl, err := t.root.New(name).Parse(in)
	if err != nil {
		return "", err
	}
	return render(tmpl, ctx)
}

// Exec - execute (render) a template - this is the built-in `template` action, except with output...
func (t *Template) Exec(name string, tmplcontext ...interface{}) (string, error) {
	ctx := t.defaultCtx
	if len(tmplcontext) == 1 {
		ctx = tmplcontext[0]
	}
	tmpl := t.root.Lookup(name)
	if tmpl == nil {
		return "", fmt.Errorf(`template "%s" not defined`, name)
	}
	return render(tmpl, ctx)
}

func render(tmpl *template.Template, ctx interface{}) (string, error) {
	out := &bytes.Buffer{}
	err := tmpl.Execute(out, ctx)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func (t *Template) parseArgs(args ...interface{}) (name, in string, ctx interface{}, err error) {
	name = "<inline>"
	ctx = t.defaultCtx

	if len(args) == 0 || len(args) > 3 {
		return "", "", nil, fmt.Errorf("wrong number of args for tpl: want 1, 2, or 3 - got %d", len(args))
	}
	first, ok := args[0].(string)
	if !ok {
		return "", "", nil, fmt.Errorf("wrong input: first arg must be string, got %T", args[0])
	}

	switch len(args) {
	case 1:
		in = first
	case 2:
		// this can either be (name string, in string) or (in string, ctx interface{})
		switch second := args[1].(type) {
		case string:
			name = first
			in = second
		default:
			in = first
			ctx = second
		}
	case 3:
		name = first
		var ok bool
		in, ok = args[1].(string)
		if !ok {
			return "", "", nil, fmt.Errorf("wrong input: second arg (in) must be string, got %T", args[0])
		}
		ctx = args[2]
	}

	return name, in, ctx, nil
}
