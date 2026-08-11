package gomplate

import (
	gocontext "context"
	"fmt"
	"regexp"
	"sync"

	"github.com/flanksource/commons/context"
	"github.com/flanksource/commons/logger"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/ext"

	"github.com/flanksource/gomplate/v3/conv"
	"github.com/flanksource/gomplate/v3/funcs"
	"github.com/flanksource/gomplate/v3/kubernetes"
	"github.com/flanksource/gomplate/v3/nilsafe"
	"github.com/flanksource/gomplate/v3/strings"
)

const celRegexProgramSizeLimit = 10_000

// staticCelEnvOptions returns the environment-independent CEL options: the
// generated functions, the kubernetes library, the cel-go extensions, the
// standard library and gomplate's own custom functions. They are identical on
// every call, so they are built once into the cached base environment (see
// baseCelEnv) instead of being reconstructed for every expression compile.
//
// Built once into the cached base env (see baseCelEnv), so this is not sized for
// speed. Starting from a nil slice also guarantees the first append copies
// funcs.CelEnvOption rather than aliasing its backing array.
func staticCelEnvOptions() []cel.EnvOption {
	var opts []cel.EnvOption //nolint:prealloc
	opts = append(opts, funcs.CelEnvOption...)
	opts = append(opts, kubernetes.Library()...)
	opts = append(opts, ext.Strings(), ext.Encoders(), ext.Lists(), ext.Math(), ext.Sets())
	opts = append(opts, cel.StdLib())
	opts = append(opts, cel.OptionalTypes(cel.OptionalTypesVersion(1)))
	opts = append(opts, nilsafe.Library(nilsafe.WithZeroValues()))
	opts = append(opts, strings.Library...)
	opts = append(opts, getGoTemplateCelFunction())
	opts = append(opts, getDebugCelFunction())
	opts = append(opts, getFoldCelLibrary())
	opts = append(opts, cel.RegexProgramSizeLimit(celRegexProgramSizeLimit))
	return opts
}

// baseCelEnv builds, exactly once, a CEL environment that holds only the static,
// environment-independent options. RunExpressionContext layers the per-call
// variables, registered native types and caller-provided functions on top with
// Env.Extend.
//
// This is the dominant CEL allocation/CPU saving: kubernetes.Library() and the
// validation of its declarations are paid once for the lifetime of the process
// instead of on every compile. EagerlyValidateDeclarations forces the base env
// to validate its declarations up front so Extend reuses them and only validates
// the small per-call delta.
//
// Env.Extend uses copy-on-write and never mutates the receiver, so the cached
// base env is safe to share across goroutines.
var baseCelEnv = sync.OnceValues(func() (*cel.Env, error) {
	opts := staticCelEnvOptions()
	opts = append(opts, cel.EagerlyValidateDeclarations(true))
	return cel.NewEnv(opts...)
})

// GetCelEnv returns the full set of CEL env options: the static options, any
// types registered via RegisterType, and one variable per environment key.
//
// RunExpressionContext no longer uses this on the hot path; it compiles against
// the cached base environment (baseCelEnv) and layers the per-call options with
// Env.Extend. GetCelEnv is retained for external callers and shares the static
// option set via staticCelEnvOptions.
func GetCelEnv(environment map[string]any) []cel.EnvOption {
	opts := staticCelEnvOptions()
	if nativeTypes := currentNativeTypes(); nativeTypes.envOption != nil {
		opts = append(opts, nativeTypes.envOption)
	}

	// Load input as variables
	for k := range environment {
		opts = append(opts, cel.Variable(k, cel.AnyType))
	}

	return opts
}

// The following identifiers are reserved to allow easier embedding of CEL into a host language.
//
// Reference: https://github.com/google/cel-spec/blob/master/doc/langdef.md
var celKeywords = map[string]struct{}{
	"true":      {},
	"false":     {},
	"null":      {},
	"in":        {},
	"as":        {},
	"break":     {},
	"const":     {},
	"continue":  {},
	"else":      {},
	"for":       {},
	"function":  {},
	"if":        {},
	"import":    {},
	"let":       {},
	"loop":      {},
	"namespace": {},
	"package":   {},
	"return":    {},
	"var":       {},
	"void":      {},
	"while":     {},
	"type":      {},
}

var celIdentifierRegexp = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// IsCelKeyword returns true if the given key is a reserved word in Cel
func IsCelKeyword(key string) bool {
	_, ok := celKeywords[key]
	return ok
}

func IsValidCELIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	return !IsCelKeyword(s) && celIdentifierRegexp.MatchString(s)
}

func getDebugCelFunction() cel.EnvOption {
	log := logger.GetLogger("cel")
	return cel.Function("debug",
		cel.Overload("debug_dyn",
			[]*cel.Type{cel.DynType},
			cel.DynType,
			cel.UnaryBinding(func(val ref.Val) ref.Val {
				log.Debugf("%s", logger.Pretty(val.Value()))
				return val
			}),
		),
		cel.Overload("debug_string_dyn",
			[]*cel.Type{cel.StringType, cel.DynType},
			cel.DynType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				log.Debugf("%s: %s", args[0].Value().(string), logger.Pretty(args[1].Value()))
				return args[1]
			}),
		),
	)
}

// getGoTemplateCelFunction returns a CEL function that calls gotemplate on a format string
func getGoTemplateCelFunction() cel.EnvOption {
	return cel.Function("f",
		cel.Overload("f_string_any",
			[]*cel.Type{
				cel.StringType, cel.DynType,
			},
			cel.StringType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				format := conv.ToString(args[0])
				data := args[1].Value()

				env := map[string]any{}
				switch v := data.(type) {
				case map[string]any:
					env = v
				case map[string]string:
					for k, v := range v {
						env[k] = v
					}
				default:
					// Otherwise, make data available as 'data' variable
					env["data"] = v
				}

				// Use struct templater as it supports ValueFunctions and multiple delims
				st := StructTemplater{
					Context:        context.NewContext(gocontext.Background()),
					Values:         env,
					ValueFunctions: true,
					DelimSets: []Delims{
						{Left: "$(", Right: ")"},
						{Left: "{{", Right: "}}"},
					},
				}
				result, err := st.Template(format)
				if err != nil {
					return types.WrapErr(fmt.Errorf("gotemplate error: %w", err))
				}

				return types.DefaultTypeAdapter.NativeToValue(result)
			}),
		),
	)
}
