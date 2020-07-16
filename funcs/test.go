package funcs

import (
	"reflect"
	"sync"

	"github.com/hairyhenderson/gomplate/v3/conv"
	"github.com/pkg/errors"

	"github.com/hairyhenderson/gomplate/v3/test"
)

var (
	testNS     *TestFuncs
	testNSInit sync.Once
)

// TestNS -
func TestNS() *TestFuncs {
	testNSInit.Do(func() { testNS = &TestFuncs{} })
	return testNS
}

// AddTestFuncs -
func AddTestFuncs(f map[string]interface{}) {
	f["test"] = TestNS

	f["assert"] = TestNS().Assert
	f["fail"] = TestNS().Fail
	f["required"] = TestNS().Required
	f["ternary"] = TestNS().Ternary
	f["kind"] = TestNS().Kind
	f["isKind"] = TestNS().IsKind
}

// TestFuncs -
type TestFuncs struct{}

// Assert -
func (f *TestFuncs) Assert(args ...interface{}) (string, error) {
	input := conv.ToBool(args[len(args)-1])
	switch len(args) {
	case 1:
		return test.Assert(input, "")
	case 2:
		message, ok := args[0].(string)
		if !ok {
			return "", errors.Errorf("at <1>: expected string; found %T", args[0])
		}
		return test.Assert(input, message)
	default:
		return "", errors.Errorf("wrong number of args: want 1 or 2, got %d", len(args))
	}
}

// Fail -
func (f *TestFuncs) Fail(args ...interface{}) (string, error) {
	switch len(args) {
	case 0:
		return "", test.Fail("")
	case 1:
		return "", test.Fail(conv.ToString(args[0]))
	default:
		return "", errors.Errorf("wrong number of args: want 0 or 1, got %d", len(args))
	}
}

// Required -
func (f *TestFuncs) Required(args ...interface{}) (interface{}, error) {
	switch len(args) {
	case 1:
		return test.Required("", args[0])
	case 2:
		message, ok := args[0].(string)
		if !ok {
			return nil, errors.Errorf("at <1>: expected string; found %T", args[0])
		}
		return test.Required(message, args[1])
	default:
		return nil, errors.Errorf("wrong number of args: want 1 or 2, got %d", len(args))
	}
}

// Ternary -
func (f *TestFuncs) Ternary(tval, fval, b interface{}) interface{} {
	if conv.ToBool(b) {
		return tval
	}
	return fval
}

// Kind - return the kind of the argument
func (f *TestFuncs) Kind(arg interface{}) string {
	return reflect.ValueOf(arg).Kind().String()
}

// IsKind - return whether or not the argument is of the given kind
func (f *TestFuncs) IsKind(kind string, arg interface{}) bool {
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
