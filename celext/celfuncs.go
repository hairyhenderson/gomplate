package celext

import (
	"encoding/json"
	"reflect"

	"github.com/flanksource/gomplate/v3/funcs"
	"github.com/flanksource/gomplate/v3/k8s"
	pkgStrings "github.com/flanksource/gomplate/v3/strings"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/ext"
)

var customCelFuncs = []cel.EnvOption{
	k8sHealth(),
	k8sIsHealthy(),
	toJSONArray(),
	toJSON(),
}

func GetCelEnv(environment map[string]any) []cel.EnvOption {
	opts := funcs.CelEnvOption

	// Generated functions
	opts = append(opts, funcs.CelEnvOption...)

	opts = append(opts, pkgStrings.CelEnvOption...)

	// load other cel-go extensions that aren't available by default
	extensions := []cel.EnvOption{ext.Math(), ext.Encoders(), ext.Strings(), ext.Sets(), ext.Lists()}
	opts = append(opts, extensions...)

	// Load input as variables
	for k := range environment {
		opts = append(opts, cel.Variable(k, cel.AnyType))
	}

	opts = append(opts, customCelFuncs...)
	return opts
}

func k8sHealth() cel.EnvOption {
	return cel.Function("k8s.health",
		cel.Overload("k8s.health_any",
			[]*cel.Type{cel.AnyType},
			cel.AnyType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				jsonObj, _ := anyToMapStringAny(k8s.GetHealth(obj.Value()))
				return types.NewDynamicMap(types.DefaultTypeAdapter, jsonObj)
			}),
		),
	)
}

func k8sIsHealthy() cel.EnvOption {
	return cel.Function("k8s.is_healthy",
		cel.Overload("k8s.is_healthy_any",
			[]*cel.Type{cel.AnyType},
			cel.StringType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				return types.Bool(k8s.GetHealth(obj.Value()).OK)
			}),
		),
	)
}

func toJSONArray() cel.EnvOption {
	return cel.Function("toJSONArray",
		cel.MemberOverload("dyn_toJSONArray_string",
			[]*cel.Type{cel.DynType},
			cel.StringType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				nativeType, err := obj.ConvertToNative(reflect.TypeOf([]map[string]any{}))
				if err != nil {
					return types.String(err.Error())
				}
				jsonStr, err := json.Marshal(nativeType)
				if err != nil {
					return types.String(err.Error())
				}
				return types.String(string(jsonStr))
			}),
		),
	)
}

func toJSON() cel.EnvOption {
	return cel.Function("toJSON",
		cel.MemberOverload("dyn_toJSON_string",
			[]*cel.Type{cel.DynType},
			cel.StringType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				nativeType, err := obj.ConvertToNative(reflect.TypeOf(map[string]any{}))
				if err != nil {
					return types.String(err.Error())
				}
				jsonStr, err := json.Marshal(nativeType)
				if err != nil {
					return types.String(err.Error())
				}
				return types.String(string(jsonStr))
			}),
		),
	)
}

func anyToMapStringAny(v any) (map[string]any, error) {
	var jsonObj map[string]any
	b, err := json.Marshal(v)
	if err != nil {
		return jsonObj, err
	}
	err = json.Unmarshal(b, &jsonObj)
	return jsonObj, err
}
