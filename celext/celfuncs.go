package celext

import (
	"encoding/json"
	"reflect"

	"github.com/flanksource/gomplate/v3/funcs"
	pkgStrings "github.com/flanksource/gomplate/v3/strings"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/ext"
	"google.golang.org/protobuf/types/known/structpb"
)

var (
	listType   = reflect.TypeOf(&structpb.ListValue{})
	mapType    = reflect.TypeOf(&structpb.Struct{})
	mapStrDyn  = cel.MapType(cel.StringType, cel.DynType)
	listStrDyn = cel.ListType(cel.DynType)
)

var customCelFuncs = []cel.EnvOption{
	k8sHealth(),
	k8sIsHealthy(),
	k8sCPUAsMillicores(),
	k8sMemoryAsBytes(),
	marshalJSON(),
	parseJSON(),
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

func anyToMapStringAny(v any) (map[string]any, error) {
	var jsonObj map[string]any
	b, err := json.Marshal(v)
	if err != nil {
		return jsonObj, err
	}
	err = json.Unmarshal(b, &jsonObj)
	return jsonObj, err
}
