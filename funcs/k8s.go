package funcs

import (
	"context"
	"encoding/json"

	"github.com/flanksource/gomplate/v3/kubernetes"
)

type KubernetesFuncs struct {
}

// CreateFilePathFuncs -
func CreateKubernetesFuncs(ctx context.Context) map[string]interface{} {
	f := map[string]interface{}{}
	ns := &KubernetesFuncs{}
	f["k8s"] = func() interface{} { return ns }
	f["isHealthy"] = ns.IsHealthy
	f["isReady"] = ns.IsReady
	f["getStatus"] = ns.GetStatus
	f["getHealth"] = ns.GetHealth
	f["neat"] = ns.Neat
	return f
}

func (ns KubernetesFuncs) IsHealthy(in interface{}) bool {
	return kubernetes.IsHealthy(in)
}

func (ns KubernetesFuncs) IsReady(in interface{}) bool {
	return kubernetes.IsReady(in)
}

func (ns KubernetesFuncs) GetStatus(in interface{}) string {
	return kubernetes.GetStatus(in)
}

func (ns KubernetesFuncs) GetHealth(in interface{}) kubernetes.HealthStatus {
	return kubernetes.GetHealth(in)
}

func (ns KubernetesFuncs) GetHealthMap(in interface{}) map[string]string {
	v := kubernetes.GetHealth(in)
	b, _ := json.Marshal(v)
	var r map[string]string
	_ = json.Unmarshal(b, &r)
	return r
}

func (ns KubernetesFuncs) Neat(in string) (string, error) {
	return kubernetes.Neat(in, "same")
}
