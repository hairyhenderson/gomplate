package funcs

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hairyhenderson/gomplate/v4/conv"
	"github.com/hairyhenderson/gomplate/v4/test"
)

// CreateTestFuncs -
func CreateTestFuncs(ctx context.Context) map[string]interface{} {
	f := map[string]interface{}{}

	ns := &TestFuncs{ctx}
	f["test"] = func() interface{} { return ns }

	f["assert"] = ns.Assert
	f["fail"] = ns.Fail
	f["required"] = ns.Required
	f["ternary"] = ns.Ternary
	f["kind"] = ns.Kind
	f["isKind"] = ns.IsKind
	return f
}

// TestFuncs -
type TestFuncs struct {
	ctx context.Context
}

// Assert -
func (TestFuncs) Assert(args ...interface{}) (string, error) {
	input := conv.ToBool(args[len(args)-1])
	switch len(args) {
	case 1:
		return test.Assert(input, "")
	case 2:
		message, ok := args[0].(string)
		if !ok {
			return "", fmt.Errorf("at <1>: expected string; found %T", args[0])
		}
		return test.Assert(input, message)
	default:
		return "", fmt.Errorf("wrong number of args: want 1 or 2, got %d", len(args))
	}
}

// Fail -
func (TestFuncs) Fail(args ...interface{}) (string, error) {
	switch len(args) {
	case 0:
		return "", test.Fail("")
	case 1:
		return "", test.Fail(conv.ToString(args[0]))
	default:
		return "", fmt.Errorf("wrong number of args: want 0 or 1, got %d", len(args))
	}
}

// Required -
func (TestFuncs) Required(args ...interface{}) (interface{}, error) {
	switch len(args) {
	case 1:
		return test.Required("", args[0])
	case 2:
		message, ok := args[0].(string)
		if !ok {
			return nil, fmt.Errorf("at <1>: expected string; found %T", args[0])
		}
		return test.Required(message, args[1])
	default:
		return nil, fmt.Errorf("wrong number of args: want 1 or 2, got %d", len(args))
	}
}

// Ternary -
func (TestFuncs) Ternary(tval, fval, b interface{}) interface{} {
	if conv.ToBool(b) {
		return tval
	}
	return fval
}

// Kind - return the kind of the argument
func (TestFuncs) Kind(arg interface{}) string {
	return reflect.ValueOf(arg).Kind().String()
}

// IsKind - return whether or not the argument is of the given kind
func (f TestFuncs) IsKind(kind string, arg interface{}) bool {
	k := f.Kind(arg)
	if kind == "number" {
		switch k {
		case "int", "int8", "int16", "int32", "int64",
			"uint", "uint8", "uint16", "uint32", "uint64", "uintptr",
			"float32", "float64",
			"complex64", "complex128":
			kind = k
		}
	}
	return k == kind
}
