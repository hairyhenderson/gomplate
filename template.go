package gomplate

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"sort"
	"strings"
	gotemplate "text/template"
	"time"

	_ "github.com/flanksource/gomplate/v3/js"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/registry"
	_ "github.com/robertkrimen/otto/underscore"
)

var funcMap gotemplate.FuncMap

var (
	// keep the cache period low as lots of anonymous functions can pile up the cache.
	goTemplateCache    = cache.New(time.Hour, time.Hour)
	celExpressionCache = cache.New(time.Hour, time.Hour)
)

func init() {
	funcMap = CreateFuncs(context.Background())
}

type Template struct {
	Template   string `yaml:"template,omitempty" json:"template,omitempty"` // Go template
	JSONPath   string `yaml:"jsonPath,omitempty" json:"jsonPath,omitempty"`
	Expression string `yaml:"expr,omitempty" json:"expr,omitempty"` // A cel-go expression
	Javascript string `yaml:"javascript,omitempty" json:"javascript,omitempty"`
	RightDelim string `yaml:"-" json:"-"`
	LeftDelim  string `yaml:"-" json:"-"`

	// Pass in additional cel-env options like functions
	// that aren't simple enough to be included in Functions
	CelEnvs []cel.EnvOption `yaml:"-" json:"-"`

	// A map of functions that are accessible to cel expressions
	// and go templates.
	// NOTE: For cel expressions, the functions must be of type func() any.
	// If any other function type is used, an error will be returned.
	// Opt to CelEnvs for those cases.
	Functions map[string]any `yaml:"-" json:"-"`
}

func (t Template) CacheKey(env map[string]any) string {
	envVars := make([]string, 0, len(env)+1)
	for k := range env {
		envVars = append(envVars, k)
	}
	sort.Slice(envVars, func(i, j int) bool { return envVars[i] < envVars[j] })

	funcNames := make([]string, 0, len(t.Functions))
	for k := range t.Functions {
		funcNames = append(funcNames, k)
	}
	sort.Slice(funcNames, func(i, j int) bool { return funcNames[i] < funcNames[j] })

	funcKeys := make([]string, 0, len(t.Functions))
	for _, fnName := range funcNames {
		funcKeys = append(funcKeys, fmt.Sprintf("%d", reflect.ValueOf(t.Functions[fnName]).Pointer()))
	}

	return strings.Join(envVars, "-") + strings.Join(funcKeys, "-") + t.RightDelim + t.LeftDelim + t.Expression + t.Javascript + t.JSONPath + t.Template
}

func (t Template) IsEmpty() bool {
	return t.Template == "" && t.JSONPath == "" && t.Expression == "" && t.Javascript == ""
}

func RunExpression(_environment map[string]any, template Template) (any, error) {
	data, err := Serialize(_environment)
	if err != nil {
		return "", err
	}

	envOptions := GetCelEnv(data)
	for name, fn := range template.Functions {
		_name := name
		_fn := fn
		envOptions = append(envOptions, cel.Function(_name, cel.Overload(
			_name,
			nil,
			cel.AnyType,
			cel.FunctionBinding(func(values ...ref.Val) ref.Val {
				ogFunc, ok := _fn.(func() any)
				if !ok {
					return types.WrapErr(fmt.Errorf("%s is expected to be of type func() any", _name))
				}

				out := ogFunc()
				return types.DefaultTypeAdapter.NativeToValue(out)
			}),
		)))
	}

	envOptions = append(envOptions, template.CelEnvs...)

	var prg cel.Program
	cached, ok := celExpressionCache.Get(template.CacheKey(_environment))
	if ok {
		if cachedPrg, ok := cached.(*cel.Program); ok {
			prg = *cachedPrg
		}
	}

	if prg == nil {
		env, err := cel.NewEnv(envOptions...)
		if err != nil {
			return "", err
		}

		ast, issues := env.Compile(strings.ReplaceAll(template.Expression, "\n", " "))
		if issues != nil && issues.Err() != nil {
			return "", issues.Err()
		}

		prg, err = env.Program(ast, cel.Globals(data))
		if err != nil {
			return "", err
		}

		celExpressionCache.SetDefault(template.CacheKey(_environment), &prg)
	}

	out, _, err := prg.Eval(data)
	if err != nil {
		return nil, errors.Wrapf(err, "error evaluating expression %s: %s", template.Expression, err)
	}
	return out.Value(), nil

}

func RunTemplate(environment map[string]any, template Template) (string, error) {
	// javascript
	if template.Javascript != "" {
		vm := otto.New()
		for k, v := range environment {
			if err := vm.Set(k, v); err != nil {
				return "", fmt.Errorf("error setting %s: %w", k, err)
			}
		}

		out, err := vm.Run(template.Javascript)
		if err != nil {
			return "", fmt.Errorf("failed to run javascript: %w", err)
		}

		if s, err := out.ToString(); err != nil {
			return "", fmt.Errorf("failed to cast output to string: %w", err)
		} else {
			return s, nil
		}
	}

	// gotemplate
	if template.Template != "" {
		return goTemplate(template, environment)
	}

	// cel-go
	if template.Expression != "" {
		out, err := RunExpression(environment, template)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v", out), nil
	}

	return "", nil
}

func goTemplate(template Template, environment map[string]any) (string, error) {
	var tpl *gotemplate.Template
	cached, ok := goTemplateCache.Get(template.CacheKey(nil))
	if ok {
		if cachedTpl, ok := cached.(*gotemplate.Template); ok {
			tpl = cachedTpl
		}
	}

	if tpl == nil {
		tpl = gotemplate.New("")
		if template.LeftDelim != "" {
			tpl = tpl.Delims(template.LeftDelim, template.RightDelim)
		}

		funcs := make(map[string]any)
		for k, v := range funcMap {
			funcs[k] = v
		}
		for k, v := range template.Functions {
			funcs[k] = v
		}
		var err error
		tpl, err = tpl.Funcs(funcs).Parse(template.Template)
		if err != nil {
			return "", err
		}

		goTemplateCache.SetDefault(template.CacheKey(nil), tpl)
	}

	data, err := Serialize(environment)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("error executing template %s: %v", strings.Split(template.Template, "\n")[0], err)
	}

	return strings.TrimSpace(buf.String()), nil
}

// LoadSharedLibrary loads a shared library for Otto
func LoadSharedLibrary(source string) error {
	source = strings.TrimSpace(source)
	data, err := os.ReadFile(source)
	if err != nil {
		return fmt.Errorf("failed to read shared library %s: %s", source, err)
	}

	fmt.Printf("Loaded %s: \n%s\n", source, string(data))
	registry.Register(func() string { return string(data) })
	return nil
}
