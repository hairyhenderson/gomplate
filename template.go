package gomplate

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	gotemplate "text/template"

	_ "github.com/flanksource/gomplate/v3/js"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/registry"
	_ "github.com/robertkrimen/otto/underscore"
)

var funcMap gotemplate.FuncMap

func init() {
	funcMap = CreateFuncs(context.Background())
}

type Template struct {
	Template   string                `yaml:"template,omitempty" json:"template,omitempty"` // Go template
	JSONPath   string                `yaml:"jsonPath,omitempty" json:"jsonPath,omitempty"`
	Expression string                `yaml:"expr,omitempty" json:"expr,omitempty"` // A cel-go expression
	Javascript string                `yaml:"javascript,omitempty" json:"javascript,omitempty"`
	Functions  map[string]func() any `yaml:"-" json:"-"`
	RightDelim string                `yaml:"-" json:"-"`
	LeftDelim  string                `yaml:"-" json:"-"`
}

func (t Template) IsEmpty() bool {
	return t.Template == "" && t.JSONPath == "" && t.Expression == "" && t.Javascript == ""
}

func RunExpression(_environment map[string]any, template Template) (any, error) {

	data, err := Serialize(_environment)
	if err != nil {
		return "", err
	}

	funcs := GetCelEnv(data)
	for name, fn := range template.Functions {
		_name := name
		_fn := fn
		funcs = append(funcs, cel.Function(_name, cel.Overload(
			_name,
			nil,
			cel.AnyType,
			cel.FunctionBinding(func(values ...ref.Val) ref.Val {
				out := _fn()
				return types.DefaultTypeAdapter.NativeToValue(out)
			}),
		)))
	}

	env, err := cel.NewEnv(funcs...)
	if err != nil {
		return "", err
	}
	ast, issues := env.Compile(strings.ReplaceAll(template.Expression, "\n", " "))
	if issues != nil && issues.Err() != nil {
		return "", issues.Err()
	}

	prg, err := env.Program(ast, cel.Globals(data))
	if err != nil {
		return "", err
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
	tpl := gotemplate.New("")

	if template.LeftDelim != "" {
		tpl = tpl.Delims(template.LeftDelim, template.RightDelim)
	}

	funcs := make(map[string]any)
	if len(template.Functions) > 0 {
		for k, v := range funcMap {
			funcs[k] = v
		}
		for k, v := range template.Functions {
			funcs[k] = v
		}
	} else {
		funcs = funcMap
	}
	for k, v := range template.Functions {
		funcs[k] = v
	}
	tpl, err := tpl.Funcs(funcs).Parse(template.Template)
	if err != nil {
		return "", err
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
