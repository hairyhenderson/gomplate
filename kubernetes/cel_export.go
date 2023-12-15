package kubernetes

import "github.com/google/cel-go/cel"

func Library() []cel.EnvOption {
	return []cel.EnvOption{
		Lists(),
		URLs(),
		Regex(),
		k8sGetHealth(),
		k8sGetStatus(),
		k8sIsHealthy(),
		k8sCPUAsMillicores(),
		k8sMemoryAsBytes(),
	}
}
