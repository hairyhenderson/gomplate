package funcs

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/hairyhenderson/gomplate/conv"
	"github.com/hairyhenderson/gomplate/regexp"
)

var (
	reNS     *ReFuncs
	reNSInit sync.Once
)

// ReNS -
func ReNS() *ReFuncs {
	reNSInit.Do(func() { reNS = &ReFuncs{} })
	return reNS
}

// AddReFuncs -
func AddReFuncs(f map[string]interface{}) {
	f["regexp"] = ReNS
}

// ReFuncs -
type ReFuncs struct{}

// Find -
func (f *ReFuncs) Find(re, input interface{}) (string, error) {
	return regexp.Find(conv.ToString(re), conv.ToString(input))
}

// FindAll -
func (f *ReFuncs) FindAll(args ...interface{}) ([]string, error) {
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
func (f *ReFuncs) Match(re, input interface{}) bool {
	return regexp.Match(conv.ToString(re), conv.ToString(input))
}

// Replace -
func (f *ReFuncs) Replace(re, replacement, input interface{}) string {
	return regexp.Replace(conv.ToString(re),
		conv.ToString(replacement),
		conv.ToString(input))
}

// ReplaceLiteral -
func (f *ReFuncs) ReplaceLiteral(re, replacement, input interface{}) (string, error) {
	return regexp.ReplaceLiteral(conv.ToString(re),
		conv.ToString(replacement),
		conv.ToString(input))
}

// Split -
func (f *ReFuncs) Split(args ...interface{}) ([]string, error) {
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
