package kubernetes

import (
	"strings"

	"github.com/flanksource/gomplate/v3/conv"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

func _k8sCPUAsMillicores(objVal string) int64 {
	var cpu int64
	if strings.HasSuffix(objVal, "m") {
		cpu = conv.ToInt64(strings.ReplaceAll(objVal, "m", ""))
	} else {
		cpu = int64(conv.ToFloat64(objVal) * 1000)
	}
	return cpu
}

func k8sCPUAsMillicores() cel.EnvOption {
	return cel.Function("k8s.cpuAsMillicores",
		cel.Overload("k8s.cpuAsMillicores_string",
			[]*cel.Type{cel.StringType},
			cel.IntType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				objVal := conv.ToString(obj.Value())
				return types.Int(_k8sCPUAsMillicores(objVal))
			}),
		),
	)
}

func _k8sMemoryAsBytes(objVal string) int64 {
	objVal = strings.ToLower(objVal)

	var memory int64
	switch {
	case strings.HasSuffix(objVal, "gi"):
		memory = int64(conv.ToFloat64(strings.ReplaceAll(objVal, "gi", "")) * 1024 * 1024 * 1024)
	case strings.HasSuffix(objVal, "mi"):
		memory = int64(conv.ToFloat64(strings.ReplaceAll(objVal, "mi", "")) * 1024 * 1024)
	case strings.HasSuffix(objVal, "ki"):
		memory = int64(conv.ToFloat64(strings.ReplaceAll(objVal, "ki", "")) * 1024)
	default:
		memory = conv.ToInt64(objVal)
	}
	return memory
}

func k8sMemoryAsBytes() cel.EnvOption {
	return cel.Function("k8s.memoryAsBytes",
		cel.Overload("k8s.memoryAsBytes_string",
			[]*cel.Type{cel.StringType},
			cel.IntType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				objVal := conv.ToString(obj.Value())
				return types.Int(_k8sMemoryAsBytes(objVal))
			}),
		),
	)
}
