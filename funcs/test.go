package funcs

import (
	"sync"

	"github.com/hairyhenderson/gomplate/conv"
	"github.com/pkg/errors"

	"github.com/hairyhenderson/gomplate/test"
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
