package gomplate

import (
	"bytes"

	"github.com/pkg/errors"
)

// tpl - a template function to do inline template processing
// Can be called 4 ways:
// {{ tpl "inline template" }} - unnamed (single-use) template with default context
// {{ tpl "name" "inline template" }} - named template with default context
// {{ tpl "inline template" $foo }} - unnamed (single-use) template with given context
// {{ tpl "name" "inline template" $foo }} - named template with given context
func (g *gomplate) tpl(args ...interface{}) (string, error) {
	name, in, ctx, err := parseArgs(args...)
	if err != nil {
		return "", err
	}
	t, err := g.rootTemplate.New(name).Parse(in)
	if err != nil {
		return "", err
	}
	out := &bytes.Buffer{}
	err = t.Execute(out, ctx)
	if err != nil {
		return "", err
	}
	return out.String(), nil
}

func parseArgs(args ...interface{}) (name, in string, ctx interface{}, err error) {
	name = "<inline>"
	ctx = &context{}

	if len(args) == 0 || len(args) > 3 {
		return "", "", nil, errors.Errorf("wrong number of args for tpl: want 1, 2, or 3 - got %d", len(args))
	}
	first, ok := args[0].(string)
	if !ok {
		return "", "", nil, errors.Errorf("wrong input: first arg must be string, got %T", args[0])
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
			return "", "", nil, errors.Errorf("wrong input: second arg (in) must be string, got %T", args[0])
		}
		ctx = args[2]
	}

	return name, in, ctx, nil
}
