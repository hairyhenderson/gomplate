package funcs

import (
	"reflect"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

var mapStringString = cel.MapType(cel.StringType, cel.StringType)

func arnToMap(fnName string) cel.EnvOption {
	return cel.Function(fnName,
		cel.Overload(fnName+"_overload",
			[]*cel.Type{
				cel.StringType,
			},
			mapStringString,
			cel.UnaryBinding(func(arg ref.Val) ref.Val {

				split := strings.Split(arg.Value().(string), ":")
				m := map[string]string{
					"service":  split[2],
					"region":   split[3],
					"account":  split[4],
					"resource": split[5],
				}
				return types.DefaultTypeAdapter.NativeToValue(m)
			}),
		),
	)
}

func fromAwsMap(fnName string) cel.EnvOption {
	return cel.Function(fnName,
		cel.Overload(fnName+"_overload",
			[]*cel.Type{
				cel.ListType(cel.MapType(cel.StringType, cel.StringType)),
			},
			cel.MapType(cel.StringType, cel.StringType),
			cel.UnaryBinding(func(arg ref.Val) ref.Val {
				list, err := arg.ConvertToNative(reflect.TypeOf([]map[string]string{}))
				if err != nil {
					return types.WrapErr(err)
				}

				var out = make(map[string]string)
				for _, i := range list.([]map[string]string) {
					out[i["Name"]] = i["Value"]
				}

				return types.DefaultTypeAdapter.NativeToValue(out)
			}),
		),
	)
}
