package gomplate

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"reflect"
	"strings"
	gotemplate "text/template"

	"github.com/flanksource/gomplate/v3/funcs"
	_ "github.com/flanksource/gomplate/v3/js"
	pkgStrings "github.com/flanksource/gomplate/v3/strings"
	"github.com/flanksource/mapstructure"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
	"github.com/robertkrimen/otto"
	"github.com/robertkrimen/otto/registry"
	_ "github.com/robertkrimen/otto/underscore"
)

var funcMap gotemplate.FuncMap

func init() {
	funcMap = CreateFuncs(context.Background())
}

type Template struct {
	Template   string `yaml:"template,omitempty" json:"template,omitempty"` // Go template
	JSONPath   string `yaml:"jsonPath,omitempty" json:"jsonPath,omitempty"`
	Expression string `yaml:"expr,omitempty" json:"expr,omitempty"` // A cel-go expression
	Javascript string `yaml:"javascript,omitempty" json:"javascript,omitempty"`
}

func (t Template) IsEmpty() bool {
	return t.Template == "" && t.JSONPath == "" && t.Expression == "" && t.Javascript == ""
}

func RunTemplate(environment map[string]any, template Template) (string, error) {
	// javascript
	if template.Javascript != "" {
		vm := otto.New()
		for k, v := range environment {
			if err := vm.Set(k, v); err != nil {
				return "", fmt.Errorf("error setting %s", k)
			}
		}

		out, err := vm.Run(template.Javascript)
		if err != nil {
			return "", fmt.Errorf("failed to run javascript: %v", err)
		}

		if s, err := out.ToString(); err != nil {
			return "", fmt.Errorf("failed to cast output to string: %v", err)
		} else {
			return s, nil
		}
	}

	// gotemplate
	if template.Template != "" {
		tpl := gotemplate.New("")
		tpl, err := tpl.Funcs(funcMap).Parse(template.Template)
		if err != nil {
			return "", err
		}

		data, err := serialize(environment)
		if err != nil {
			return "", err
		}

		var buf bytes.Buffer
		if err := tpl.Execute(&buf, data); err != nil {
			return "", fmt.Errorf("error executing template %s: %v", strings.Split(template.Template, "\n")[0], err)
		}
		return strings.TrimSpace(buf.String()), nil
	}

	// cel-go
	if template.Expression != "" {
		var opts = funcs.CelEnvOption
		opts = append(opts, pkgStrings.CelEnvOption...)

		// load other cel-go extensions that aren't available by default
		extensions := []cel.EnvOption{ext.Math(), ext.Encoders(), ext.Strings(), ext.Sets(), ext.Lists()}
		opts = append(opts, extensions...)

		for k := range environment {
			opts = append(opts, cel.Variable(k, cel.AnyType))
		}

		env, err := cel.NewEnv(opts...)
		if err != nil {
			return "", err
		}

		ast, issues := env.Compile(template.Expression)
		if issues != nil && issues.Err() != nil {
			return "", issues.Err()
		}

		prg, err := env.Program(ast, cel.Globals(environment))
		if err != nil {
			return "", err
		}

		data, err := serialize(environment)
		if err != nil {
			return "", err
		}

		out, _, err := prg.Eval(data)
		if err != nil {
			return "", fmt.Errorf("error evaluating expression %s: %v", template.Expression, err)
		}

		return fmt.Sprintf("%v", out.Value()), nil
	}

	return "", nil
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

// serialize iterates over each key-value pair in the input map
// serializes any struct value to map[string]any.
func serialize(in map[string]any) (map[string]any, error) {
	if in == nil {
		return nil, nil
	}

	newMap := make(map[string]any, len(in))
	for k, v := range in {
		var dec *mapstructure.Decoder
		var err error

		vt := reflect.TypeOf(v)
		switch vt.Kind() {
		case reflect.Struct:
			var result map[string]any
			dec, err = mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "json", Result: &result, Squash: true, Deep: true})
			if err != nil {
				return nil, fmt.Errorf("error creating new mapstructure decoder: %w", err)
			}

			if err := dec.Decode(v); err != nil {
				return nil, fmt.Errorf("error decoding %T to map[string]any: %w", v, err)
			}

			newMap[k] = result

		case reflect.Slice:
			var result any
			if vt.Elem().Kind() == reflect.Struct {
				result = make([]map[string]any, 0)
			}

			dec, err = mapstructure.NewDecoder(&mapstructure.DecoderConfig{TagName: "json", Result: &result, Squash: true, Deep: true})
			if err != nil {
				return nil, fmt.Errorf("error creating new mapstructure decoder: %w", err)
			}
			if err := dec.Decode(v); err != nil {
				return nil, fmt.Errorf("error decoding %T to map[string]any: %w", v, err)
			}

			newMap[k] = result

		default:
			newMap[k] = v
			continue
		}
	}

	return newMap, nil
}
