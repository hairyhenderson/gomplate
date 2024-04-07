package kubernetes

import (
	"fmt"
	"strings"

	"github.com/flanksource/gomplate/v3/conv"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func celk8sLabels() cel.EnvOption {
	return cel.Function("k8s.labels",
		cel.Overload("k8s.labels_map_map",
			[]*cel.Type{cel.AnyType},
			cel.AnyType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				val := k8sLabels(obj.Value())
				return types.NewStringStringMap(types.DefaultTypeAdapter, val)
			}),
		),
	)
}

func k8sLabels(input any) map[string]string {
	labels := make(map[string]string)

	obj := GetUnstructured(input)
	if obj == nil {
		return labels
	}

	if ns := obj.GetNamespace(); ns != "" {
		labels["namespace"] = ns
	}

	for k, v := range obj.GetLabels() {
		if strings.HasSuffix(k, "-hash") {
			continue
		}
		labels[k] = v
	}

	return labels
}

func celPodProperties() cel.EnvOption {
	return cel.Function("k8s.podProperties",
		cel.Overload("k8s.podProperties_list_dyn_map",
			[]*cel.Type{cel.AnyType},
			cel.AnyType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				jsonObj, _ := conv.AnyToListMapStringAny(PodComponentProperties(obj.Value()))
				return types.NewDynamicList(types.DefaultTypeAdapter, jsonObj)
			}),
		),
	)
}

func celPodResourceLimits() cel.EnvOption {
	return cel.Function("k8s.getResourcesLimit",
		cel.Overload("k8s.getResourcesLimit_obj_str_int",
			[]*cel.Type{cel.AnyType, cel.StringType},
			cel.AnyType,
			cel.BinaryBinding(func(obj, resourceType ref.Val) ref.Val {
				val := getPodResources(obj.Value(), fmt.Sprint(resourceType.Value()), "limits")
				return types.Int(val)
			}),
		),
	)
}

func celPodResourceRequests() cel.EnvOption {
	return cel.Function("k8s.getResourcesRequests",
		cel.Overload("k8s.getResourcesRequests_obj_str_int",
			[]*cel.Type{cel.AnyType, cel.StringType},
			cel.AnyType,
			cel.BinaryBinding(func(obj, resourceType ref.Val) ref.Val {
				val := getPodResources(obj.Value(), fmt.Sprint(resourceType.Value()), "requests")
				return types.Int(val)
			}),
		),
	)
}

func getPodResources(input any, resourceType string, allocType string) int64 {
	obj := GetUnstructured(input)
	if obj == nil {
		return 0
	}

	var pod corev1.Pod
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &pod)
	if err != nil {
		return 0
	}

	if resourceType == "memory" && allocType == "limits" {
		var totalMemBytes int64
		for _, container := range pod.Spec.Containers {
			mem := container.Resources.Limits.Memory()
			if mem != nil {
				totalMemBytes += _k8sMemoryAsBytes(mem.String())
			}
		}
		return totalMemBytes
	}

	if resourceType == "memory" && allocType == "requests" {
		var totalMemBytes int64
		for _, container := range pod.Spec.Containers {
			mem := container.Resources.Requests.Memory()
			if mem != nil {
				totalMemBytes += _k8sMemoryAsBytes(mem.String())
			}
		}
		return totalMemBytes
	}

	if resourceType == "cpu" && allocType == "limits" {
		var totalCPU int64
		for _, container := range pod.Spec.Containers {
			cpu := container.Resources.Limits.Cpu()
			if cpu != nil {
				totalCPU += _k8sCPUAsMillicores(cpu.String())
			}
		}
		return totalCPU

	}

	if resourceType == "cpu" && allocType == "requests" {
		var totalCPU int64
		for _, container := range pod.Spec.Containers {
			cpu := container.Resources.Requests.Cpu()
			if cpu != nil {
				totalCPU += _k8sCPUAsMillicores(cpu.String())
			}
		}
		return totalCPU
	}

	return 0
}

func PodComponentProperties(input any) []map[string]any {
	obj := GetUnstructured(input)
	if obj == nil {
		return nil
	}

	var pod corev1.Pod
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &pod)
	if err != nil {
		return nil
	}

	var totalCPU int64
	for _, container := range pod.Spec.Containers {
		cpu := container.Resources.Limits.Cpu()
		if cpu != nil {
			totalCPU += _k8sCPUAsMillicores(cpu.String())
		}
	}

	var totalMemBytes int64
	for _, container := range pod.Spec.Containers {
		mem := container.Resources.Limits.Memory()
		if mem != nil {
			totalMemBytes += _k8sMemoryAsBytes(mem.String())
		}
	}

	rootContainer := pod.Spec.Containers[0]
	return []map[string]any{
		{"name": "image", "text": rootContainer.Image},
		{"name": "cpu", "max": totalCPU, "unit": "millicores", "headline": true},
		{"name": "memory", "max": totalMemBytes, "unit": "bytes", "headline": true},
		{"name": "node", "text": pod.Spec.NodeName},
		{"name": "created_at", "text": pod.ObjectMeta.CreationTimestamp.String()},
		{"name": "namespace", "text": pod.ObjectMeta.Namespace},
	}
}

func celNodeProperties() cel.EnvOption {
	return cel.Function("k8s.nodeProperties",
		cel.Overload("k8s.nodeProperties_list_dyn_map",
			[]*cel.Type{cel.AnyType},
			cel.AnyType,
			cel.UnaryBinding(func(obj ref.Val) ref.Val {
				jsonObj, _ := conv.AnyToListMapStringAny(NodeComponentProperties(obj.Value()))
				return types.NewDynamicList(types.DefaultTypeAdapter, jsonObj)
			}),
		),
	)
}

func NodeComponentProperties(input any) []map[string]any {
	obj := GetUnstructured(input)
	if obj == nil {
		return nil
	}

	var node corev1.Node
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &node)
	if err != nil {
		return nil
	}

	totalCPU := _k8sCPUAsMillicores(node.Status.Allocatable.Cpu().String())
	totalMemBytes := _k8sMemoryAsBytes(node.Status.Allocatable.Memory().String())
	totalStorage := _k8sMemoryAsBytes(node.Status.Allocatable.StorageEphemeral().String())

	return []map[string]any{
		{"name": "cpu", "max": totalCPU, "unit": "millicores", "headline": true},
		{"name": "memory", "max": totalMemBytes, "unit": "bytes", "headline": true},
		{"name": "ephemeral-storage", "max": totalStorage, "unit": "bytes", "headline": true},
		{"name": "zone", "text": node.GetLabels()["topology.kubernetes.io/zone"]},
	}
}
