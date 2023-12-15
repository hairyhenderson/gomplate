package kubernetes

import (
	"encoding/json"
	"fmt"

	"github.com/flanksource/gomplate/v3/conv"
	"github.com/flanksource/is-healthy/pkg/health"
	"github.com/flanksource/is-healthy/pkg/lua"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

type HealthStatus struct {
	Status  string `json:"status"`
	Message string `json:"message"`
	OK      bool   `json:"ok"`
}

func GetUnstructuredMap(in interface{}) []byte {
	x := GetUnstructured(in)
	b, _ := json.Marshal(x)
	return b
}

func GetUnstructured(in interface{}) *unstructured.Unstructured {
	var err error
	obj := make(map[string]interface{})

	switch v := in.(type) {
	case string:
		err = yaml.Unmarshal([]byte(v), &obj)
	case []byte:
		err = yaml.Unmarshal(v, &obj)
	case types.Bytes:
		err = yaml.Unmarshal(v, &obj)
	case map[string]interface{}:
		if val, ok := v["Object"].(map[string]any); ok {
			obj = val
		} else {
			obj = v
		}
	case unstructured.Unstructured:
		obj = v.Object
	default:
		var data []byte
		if data, err = yaml.Marshal(in); err == nil {
			err = yaml.Unmarshal(data, &obj)
		}
	}

	if err != nil {
		return nil
	}

	return &unstructured.Unstructured{Object: obj}
}

func IsHealthy(in interface{}) bool {
	return GetHealth(in).OK
}

func GetStatus(in interface{}) string {
	health := GetHealth(in)
	return fmt.Sprintf("%s: %s", health.Status, health.Message)
}

func GetHealth(in interface{}) HealthStatus {
	var err error
	obj := GetUnstructured(in)

	if obj == nil {
		return HealthStatus{
			OK:      false,
			Status:  "Error",
			Message: "Invalid spec",
		}
	}

	_health, err := health.GetResourceHealth(obj, lua.ResourceHealthOverrides{})
	if err != nil {
		return HealthStatus{
			OK:      false,
			Status:  "Error",
			Message: err.Error(),
		}
	}

	if _health == nil {
		return HealthStatus{
			OK:      false,
			Status:  "Missing",
			Message: "No health check found",
		}
	}

	return HealthStatus{
		OK:      _health.Status == health.HealthStatusHealthy || _health.Status == health.HealthStatusProgressing,
		Status:  string(_health.Status),
		Message: _health.Message,
	}
}

func k8sGetHealth() cel.EnvOption {
	return cel.Function("k8s.getHealth",
		cel.Overload("GetHealth_any",
			[]*cel.Type{cel.AnyType},
			cel.AnyType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				jsonObj, _ := conv.AnyToMapStringAny(GetHealth(obj.Value()))
				return types.NewDynamicMap(types.DefaultTypeAdapter, jsonObj)
			}),
		),
	)
}

func k8sGetStatus() cel.EnvOption {
	return cel.Function("k8s.getStatus",
		cel.Overload("GetStatus",
			[]*cel.Type{cel.AnyType},
			cel.AnyType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				return types.String(GetStatus(obj.Value()))
			}),
		),
	)
}

func k8sIsHealthy() cel.EnvOption {
	return cel.Function("k8s.isHealthy",
		cel.Overload("IsHealthy_interface{}",
			[]*cel.Type{cel.AnyType},
			cel.StringType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				return types.Bool(GetHealth(obj.Value()).OK)
			}),
		),
	)
}
