package funcs

import (
	"context"

	"github.com/pkg/errors"

	"github.com/flanksource/gomplate/v3/conv"
	"github.com/flanksource/gomplate/v3/regexp"
)

// ReNS -
// Deprecated: don't use
func ReNS() *ReFuncs {
	return &ReFuncs{}
}

// AddReFuncs -
// Deprecated: use CreateReFuncs instead
func AddReFuncs(f map[string]interface{}) {
	for k, v := range CreateReFuncs(context.Background()) {
		f[k] = v
	}
}

// CreateReFuncs -
func CreateReFuncs(ctx context.Context) map[string]interface{} {
	ns := &ReFuncs{ctx}
	return map[string]interface{}{
		"regexp": func() interface{} { return ns },
	}
}

// ReFuncs -
type ReFuncs struct {
	ctx context.Context
}

// Find -
func (ReFuncs) Find(re, input interface{}) (string, error) {
	return regexp.Find(conv.ToString(re), conv.ToString(input))
}

// FindAll -
func (ReFuncs) FindAll(args ...interface{}) ([]string, error) {
	re := ""
	n := 0
	input := ""
	switch len(args) {
	case 2:
		n = -1
		re = conv.ToString(args[0])
		input = conv.ToString(args[1])
	case 3:
		re = conv.ToString(args[0])
		n = conv.ToInt(args[1])
		input = conv.ToString(args[2])
	default:
		return nil, errors.Errorf("wrong number of args: want 2 or 3, got %d", len(args))
	}
	return regexp.FindAll(re, n, input)
}

// Match -
func (ReFuncs) Match(re, input interface{}) bool {
	return regexp.Match(conv.ToString(re), conv.ToString(input))
}

// QuoteMeta -
func (ReFuncs) QuoteMeta(in interface{}) string {
	return regexp.QuoteMeta(conv.ToString(in))
}

// Replace -
func (ReFuncs) Replace(re, replacement, input interface{}) string {
	return regexp.Replace(conv.ToString(re),
		conv.ToString(replacement),
		conv.ToString(input))
}

// ReplaceLiteral -
func (ReFuncs) ReplaceLiteral(re, replacement, input interface{}) (string, error) {
	return regexp.ReplaceLiteral(conv.ToString(re),
		conv.ToString(replacement),
		conv.ToString(input))
}

// Split -
func (ReFuncs) Split(args ...interface{}) ([]string, error) {
	re := ""
	n := -1
	input := ""
	switch len(args) {
	case 2:
		re = conv.ToString(args[0])
		input = conv.ToString(args[1])
	case 3:
		re = conv.ToString(args[0])
		n = conv.ToInt(args[1])
		input = conv.ToString(args[2])
	default:
		return nil, errors.Errorf("wrong number of args: want 2 or 3, got %d", len(args))
	}
	return regexp.Split(re, n, input)
}
