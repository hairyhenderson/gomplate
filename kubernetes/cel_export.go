package kubernetes

import "github.com/google/cel-go/cel"

func Library() []cel.EnvOption {
	return []cel.EnvOption{
		Lists(),
		URLs(),
		Regex(),
		k8sNeat(), k8sNeatWithOption(),
		k8sGetHealth("k8s.getHealth"), k8sGetHealth("GetHealth"),
		k8sGetStatus("k8s.getStatus"), k8sGetStatus("GetStatus"),
		k8sIsHealthy("k8s.isHealthy"), k8sIsHealthy("IsHealthy"), k8sIsHealthy("k8s.is_healthy"),
		k8sCPUAsMillicores(),
		k8sMemoryAsBytes(),
		celPodProperties(),
		celNodeProperties(),
		celk8sLabels(),
		celPodResourceLimits(),
		celPodResourceRequests(),
	}
}
