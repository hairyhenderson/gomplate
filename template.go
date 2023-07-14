package gomplate

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	gotemplate "text/template"

	"github.com/flanksource/gomplate/v3/funcs"
	"github.com/google/cel-go/cel"
)

var funcMap gotemplate.FuncMap

func init() {
	funcMap = CreateFuncs(context.Background())
}

type Template struct {
	Template   string `yaml:"template,omitempty" json:"template,omitempty"`
	JSONPath   string `yaml:"jsonPath,omitempty" json:"jsonPath,omitempty"`
	Expression string `yaml:"expr,omitempty" json:"expr,omitempty"`
	Javascript string `yaml:"javascript,omitempty" json:"javascript,omitempty"`
}

func (t Template) IsEmpty() bool {
	return t.Template == "" && t.JSONPath == "" && t.Expression == "" && t.Javascript == ""
}

func RunTemplate(environment map[string]interface{}, template Template) (string, error) {
	// javascript
	if template.Javascript != "" {
		// // FIXME: whitelist allowed files
		// vm := otto.New()
		// for k, v := range environment {
		// 	if err := vm.Set(k, v); err != nil {
		// 		return "", errors.Wrapf(err, "error setting %s", k)
		// 	}
		// }

		// if err != nil {
		// 	return "", errors.Wrapf(err, "error setting findConfigItem function")
		// }

		// out, err := vm.Run(template.Javascript)
		// if err != nil {
		// 	return "", errors.Wrapf(err, "failed to run javascript")
		// }

		// if s, err := out.ToString(); err != nil {
		// 	return "", errors.Wrapf(err, "failed to cast output to string")
		// } else {
		// 	return s, nil
		// }
	}

	// gotemplate
	if template.Template != "" {
		tpl := gotemplate.New("")
		tpl, err := tpl.Funcs(funcMap).Parse(template.Template)
		if err != nil {
			return "", err
		}

		// marshal data from interface{} to map[string]interface{}
		data, _ := json.Marshal(environment)
		unstructured := make(map[string]interface{})
		if err := json.Unmarshal(data, &unstructured); err != nil {
			return "", err
		}

		var buf bytes.Buffer
		if err := tpl.Execute(&buf, unstructured); err != nil {
			return "", fmt.Errorf("error executing template %s: %v", strings.Split(template.Template, "\n")[0], err)
		}
		return strings.TrimSpace(buf.String()), nil
	}

	// cel-go
	if template.Expression != "" {
		var opts = funcs.CelEnvOption
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
		out, _, err := prg.Eval(environment)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%v", out.Value()), nil
	}

	return "", nil
}
