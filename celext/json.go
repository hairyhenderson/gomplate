package celext

import (
	"encoding/json"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// Reference: https://github.com/tektoncd/triggers/blob/main/pkg/interceptors/cel/triggers.go

func toJSON() cel.EnvOption {
	valToJSONString := func(val ref.Val) ref.Val {
		var typeDesc reflect.Type

		switch val.Type() {
		case types.MapType:
			typeDesc = mapType
		case types.ListType:
			typeDesc = listType
		default:
			return types.ValOrErr(val, "unexpected type:%v passed to toJSON", val.Type())
		}

		nativeVal, err := val.ConvertToNative(typeDesc)
		if err != nil {
			return types.NewErr("failed to convert to native: %w", err)
		}

		marshaledVal, err := json.Marshal(nativeVal)
		if err != nil {
			return types.NewErr("failed to marshal to json: %w", err)
		}

		return types.String(marshaledVal)
	}

	return cel.Function("toJSON",
		cel.MemberOverload("toJSON_dyn", []*cel.Type{cel.DynType}, cel.StringType,
			cel.UnaryBinding(valToJSONString)),
	)
}

func parseJSON() cel.EnvOption {
	parseJSONString := func(val ref.Val) ref.Val {
		str := val.(types.String)
		decodedVal := map[string]interface{}{}
		err := json.Unmarshal([]byte(str), &decodedVal)
		if err != nil {
			return types.NewErr("failed to decode '%v' in parseJSON: %w", str, err)
		}
		r, err := types.NewRegistry()
		if err != nil {
			return types.NewErr("failed to create a new registry in parseJSON: %w", err)
		}
		return types.NewDynamicMap(r, decodedVal)
	}

	return cel.Function("parseJSON",
		cel.MemberOverload("parseJSON_string", []*cel.Type{cel.StringType}, mapStrDyn,
			cel.UnaryBinding(parseJSONString)),
	)
}
