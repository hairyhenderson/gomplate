package gomplate

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	commonsContext "github.com/flanksource/commons/context"
	"github.com/flanksource/commons/properties"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/patrickmn/go-cache"
	"github.com/samber/oops"
)

var celExpressionCache = cache.New(time.Hour, time.Hour)

func RunExpression(environment map[string]any, template Template) (any, error) {
	return RunExpressionContext(newContext(), environment, template)
}

func RunExpressionContext(ctx commonsContext.Context, environment map[string]any, template Template) (any, error) {
	tracker := celTrackerFromContext(ctx)
	if tracker != nil {
		if err := tracker.begin(); err != nil {
			return nil, err
		}
		defer tracker.abort()
	}

	nativeTypes := currentNativeTypes()
	data, err := serializeForCEL(environment, nativeTypes)
	if err != nil {
		return "", err
	}
	cacheKey := template.celCacheKey(environment, nativeTypes.generation)

	var program cel.Program
	var ast *cel.Ast
	if tracker == nil && template.IsCacheable() {
		cached, found := celExpressionCache.Get(cacheKey)
		if found {
			if cachedProgram, ok := cached.(*cel.Program); ok {
				program = *cachedProgram
			}
		}
	}

	if program == nil {
		program, ast, err = compileCELProgram(data, template, nativeTypes, tracker != nil)
		if err != nil {
			return "", err
		}
		if tracker == nil && template.IsCacheable() {
			celExpressionCache.Set(cacheKey, &program, template.CacheTime)
		}
	}

	out, details, err := program.Eval(data)
	if tracker != nil {
		tracker.complete(ast, details, out)
	}
	if err != nil {
		return nil, oops.With("template", template.Expression).Wrap(err)
	}
	if ctx.Logger != nil && out.Value() != template.Expression && properties.On(false, "gomplate.log") {
		ctx.Logger.V(4).Infof("templated %s => %v", template.ShortString(), out)
	}
	return out.Value(), nil
}

func compileCELProgram(data map[string]any, template Template, nativeTypes *nativeTypeSnapshot, trackState bool) (cel.Program, *cel.Ast, error) {
	base, err := baseCelEnv()
	if err != nil {
		return nil, nil, err
	}

	envOptions := celEnvOptions(data, template, nativeTypes)
	env, err := base.Extend(envOptions...)
	if err != nil {
		return nil, nil, err
	}
	expression := strings.ReplaceAll(template.Expression, "\n", " ")
	if trackState {
		expression = template.Expression
	}
	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, nil, oops.With("template", template.Expression).Errorf("issues: %s", issues.String())
	}

	evalOptions := []cel.EvalOption{cel.OptOptimize}
	if trackState {
		evalOptions = append(evalOptions, cel.OptTrackState)
	}
	program, err := env.Program(ast, cel.EvalOptions(evalOptions...))
	if err != nil {
		return nil, nil, err
	}
	return program, ast, nil
}

func celEnvOptions(data map[string]any, template Template, nativeTypes *nativeTypeSnapshot) []cel.EnvOption {
	envOptions := make([]cel.EnvOption, 0, len(data)+len(template.Functions)+len(template.CelEnvs)+1)
	if nativeTypes.envOption != nil {
		envOptions = append(envOptions, nativeTypes.envOption)
	}
	for key := range data {
		envOptions = append(envOptions, cel.Variable(key, cel.AnyType))
	}
	for name, function := range template.Functions {
		functionName := name
		registeredFunction := function
		envOptions = append(envOptions, cel.Function(functionName, cel.Overload(
			functionName,
			nil,
			cel.AnyType,
			cel.FunctionBinding(func(_ ...ref.Val) ref.Val {
				function, ok := registeredFunction.(func() any)
				if !ok {
					return types.WrapErr(fmt.Errorf("%s is expected to be of type func() any", functionName))
				}
				return types.DefaultTypeAdapter.NativeToValue(function())
			}),
		)))
	}
	envOptions = append(envOptions, template.CelEnvs...)
	return envOptions
}

func (t Template) celCacheKey(environment map[string]any, nativeTypeGeneration uint64) string {
	if nativeTypeGeneration == 0 {
		return t.cacheKey(environment)
	}
	return strconv.FormatUint(nativeTypeGeneration, 10) + ":" + t.cacheKey(environment)
}
