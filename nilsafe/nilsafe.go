package nilsafe

import (
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/operators"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/interpreter"
)

type Option func(*config)

type config struct {
	zeroValues bool
}

func WithZeroValues() Option { return func(c *config) { c.zeroValues = true } }

// Library returns a cel.EnvOption that enables nil-safe evaluation.
// Missing variables, null field access, missing map keys, and out-of-bounds
// list indices return null instead of errors. Real errors like division by zero,
// type conversion failures, and regex errors are preserved.
func Library(opts ...Option) cel.EnvOption {
	cfg := &config{}
	for _, o := range opts {
		o(cfg)
	}
	return cel.Lib(&library{cfg: cfg})
}

type library struct {
	cfg *config
}

func (*library) LibraryName() string             { return "cel.lib.ext.nilsafe" }
func (*library) CompileOptions() []cel.EnvOption { return nil }

func (l *library) ProgramOptions() []cel.ProgramOption {
	return []cel.ProgramOption{cel.CustomDecoratorV2(l.makeDecorator())}
}

func (l *library) makeDecorator() interpreter.InterpretableDecoratorV2 {
	return func(i interpreter.InterpretableV2) (interpreter.InterpretableV2, error) {
		if attr, ok := i.(interpreter.InterpretableAttribute); ok {
			if attr.ID() != attr.Attr().ID() {
				return i, nil
			}
			return &nilSafeAttr{InterpretableAttribute: attr}, nil
		}
		if call, ok := i.(interpreter.InterpretableCall); ok {
			fn := call.Function()
			if fn == "coalesce" || fn == "first" || fn == "last" {
				return i, nil
			}
			if l.cfg.zeroValues {
				if fn == operators.Equals {
					return &zeroValueEq{InterpretableCall: call}, nil
				}
				if fn == operators.NotEquals {
					return &zeroValueNe{InterpretableCall: call}, nil
				}
				return &zeroValueCall{InterpretableCall: call}, nil
			}
			if fn == operators.Equals || fn == operators.NotEquals {
				return i, nil
			}
			return &nilSafeCall{InterpretableCall: call}, nil
		}
		return i, nil
	}
}

type nilSafeAttr struct {
	interpreter.InterpretableAttribute
}

func (a *nilSafeAttr) Eval(ctx interpreter.Activation) ref.Val {
	return nilSafeResolution(a.InterpretableAttribute.Eval(ctx))
}

func (a *nilSafeAttr) Exec(frame *interpreter.ExecutionFrame) ref.Val {
	return nilSafeResolution(a.InterpretableAttribute.Exec(frame))
}

func nilSafeResolution(val ref.Val) ref.Val {
	if types.IsError(val) && isResolutionError(val) {
		return types.NullValue
	}
	return val
}

func isResolutionError(val ref.Val) bool {
	e, ok := val.(*types.Err)
	if !ok {
		return false
	}
	msg := e.Unwrap().Error()
	return strings.Contains(msg, "no such attribute") ||
		strings.Contains(msg, "no such key") ||
		strings.Contains(msg, "index out of bounds")
}

type nilSafeCall struct {
	interpreter.InterpretableCall
}

func (c *nilSafeCall) Eval(ctx interpreter.Activation) ref.Val {
	return c.eval(
		func(arg interpreter.InterpretableV2) ref.Val { return arg.Eval(ctx) },
		func() ref.Val { return c.InterpretableCall.Eval(ctx) },
	)
}

func (c *nilSafeCall) Exec(frame *interpreter.ExecutionFrame) ref.Val {
	return c.eval(
		func(arg interpreter.InterpretableV2) ref.Val { return arg.Exec(frame) },
		func() ref.Val { return c.InterpretableCall.Exec(frame) },
	)
}

func (c *nilSafeCall) eval(evalArg func(interpreter.InterpretableV2) ref.Val, evalCall func() ref.Val) ref.Val {
	for _, arg := range c.Args() {
		if evalArg(arg) == types.NullValue {
			return types.NullValue
		}
	}
	return evalCall()
}
