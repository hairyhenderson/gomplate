package gomplate

import (
	"github.com/flanksource/gomplate/v3/funcs"
	"github.com/flanksource/gomplate/v3/kubernetes"
	"github.com/flanksource/gomplate/v3/strings"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
)

func GetCelEnv(environment map[string]any) []cel.EnvOption {
	// Generated functions
	var opts = funcs.CelEnvOption
	opts = append(opts, kubernetes.Library()...)
	opts = append(opts, ext.Strings(), ext.Encoders(), ext.Lists(), ext.Math(), ext.Sets(), cel.StdLib(), cel.OptionalTypes())
	opts = append(opts, strings.Library...)

	// Load input as variables
	for k := range environment {
		opts = append(opts, cel.Variable(k, cel.AnyType))
	}

	return opts
}
