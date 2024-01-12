package kubernetes

import (
	"github.com/flanksource/gomplate/v3/conv"
	"github.com/flanksource/kubectl-neat/cmd"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

func Neat(in, outputFormat string) (string, error) {
	out, err := cmd.NeatYAMLOrJSON([]byte(in), outputFormat)
	return string(out), err
}

func k8sNeat() cel.EnvOption {
	return cel.Function("k8s.neat",
		cel.Overload("k8s_neat",
			[]*cel.Type{cel.StringType},
			cel.StringType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				objVal := conv.ToString(obj.Value())
				res, err := Neat(objVal, "same")
				if err != nil {
					return types.NewErr(err.Error())
				}

				return types.String(res)
			}),
		),
	)
}

func k8sNeatWithOption() cel.EnvOption {
	return cel.Function("k8s.neat",
		cel.Overload("k8s_neat_with_option",
			[]*cel.Type{cel.StringType, cel.StringType},
			cel.StringType,
			cel.FunctionBinding(func(args ...ref.Val) ref.Val {
				objVal := conv.ToString(args[0].Value())
				outputFormat := conv.ToString(args[1].Value())

				res, err := Neat(objVal, outputFormat)
				if err != nil {
					return types.NewErr(err.Error())
				}

				return types.String(res)
			}),
		),
	)
}
