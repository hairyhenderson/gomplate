package funcs

import (
	"context"
	"encoding/json"

	"github.com/flanksource/gomplate/v3/k8s"
)

type KubernetesFuncs struct {
}

// CreateFilePathFuncs -
func CreateKubernetesFuncs(ctx context.Context) map[string]interface{} {
	f := map[string]interface{}{}
	ns := &KubernetesFuncs{}
	f["k8s"] = func() interface{} { return ns }
	f["isHealthy"] = ns.IsHealthy
	f["getStatus"] = ns.GetStatus
	f["getHealth"] = ns.GetHealth
	return f
}

func (ns KubernetesFuncs) IsHealthy(in interface{}) bool {
	return k8s.IsHealthy(in)
}

func (ns KubernetesFuncs) GetStatus(in interface{}) string {
	return k8s.GetStatus(in)
}

func (ns KubernetesFuncs) GetHealth(in interface{}) k8s.HealthStatus {
	return k8s.GetHealth(in)
}

func (ns KubernetesFuncs) GetHealthMap(in interface{}) map[string]string {
	v := k8s.GetHealth(in)
	b, _ := json.Marshal(v)
	var r map[string]string
	_ = json.Unmarshal(b, &r)
	return r
}
