package gomplate

import (
	"testing"
	"time"

	"github.com/flanksource/gomplate/v3/k8s"
	"github.com/stretchr/testify/assert"
)

func TestGomplate(t *testing.T) {
	tests := []struct {
		env      map[string]interface{}
		template string
		out      string
	}{
		{map[string]interface{}{"hello": "world"}, "{{ .hello }}", "world"},
		{map[string]interface{}{"age": 75 * time.Second}, "{{ .age | humanDuration  }}", "1m15s"},
		{map[string]interface{}{"healthySvc": k8s.GetUnstructured(k8s.TestHealthy)}, "{{ (.healthySvc | isHealthy) }}", "true"},
		{map[string]interface{}{"healthySvc": k8s.GetUnstructured(k8s.TestLuaStatus)}, "{{ (.healthySvc | getStatus) }}", "Degraded: found less than two generators, Merge requires two or more"},
		{map[string]interface{}{"healthySvc": k8s.GetUnstructured(k8s.TestHealthy)}, "{{ (.healthySvc | getHealth).Status  }}", "Healthy"},
		{map[string]interface{}{"size": 123456}, "{{ .size | humanSize }}", "120.6K"},
		{map[string]interface{}{"v": "1.2.3-beta.1+c0ff33"}, "{{  (.v | semver).Prerelease  }}", "beta.1"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.3"}, "{{  .old | semverCompare .new }}", "true"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.4"}, "{{  .old | semverCompare .new }}", "false"},
	}

	for _, tc := range tests {
		t.Run(tc.template, func(t *testing.T) {
			out, err := RunTemplate(tc.env, Template{
				Template: tc.template,
			})
			assert.ErrorIs(t, err, nil)
			assert.Equal(t, tc.out, out)
		})
	}
}

func TestCel(t *testing.T) {
	tests := []struct {
		env        map[string]interface{}
		expression string
		out        string
	}{
		{nil, `math.Add([1,2,3,4,5])`, "15"},
		{map[string]interface{}{"hello": "world"}, "hello", "world"},
		{map[string]interface{}{"age": 75 * time.Second}, "age", "1m15s"},
		{map[string]interface{}{"healthySvc": k8s.GetUnstructuredMap(k8s.TestHealthy)}, "IsHealthy(healthySvc)", "true"},
		{map[string]interface{}{"healthySvc": k8s.GetUnstructuredMap(k8s.TestLuaStatus)}, "GetStatus(healthySvc)", "Degraded: found less than two generators, Merge requires two or more"},
		{map[string]interface{}{"healthySvc": k8s.GetUnstructuredMap(k8s.TestHealthy)}, "GetHealth(healthySvc).status", "Healthy"},
		{map[string]interface{}{"size": "123456"}, "HumanSize(size)", "120.6K"},
		{map[string]interface{}{"size": 123456}, "HumanSize(size)", "120.6K"},
		{map[string]interface{}{"v": "1.2.3-beta.1+c0ff33"}, "Semver(v).prerelease", "beta.1"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.3"}, "SemverCompare(new, old)", "true"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.4"}, "SemverCompare(new, old)", "false"},
	}

	for _, tc := range tests {
		t.Run(tc.expression, func(t *testing.T) {
			out, err := RunTemplate(tc.env, Template{
				Expression: tc.expression,
			})
			assert.ErrorIs(t, nil, err)
			assert.Equal(t, tc.out, out)
		})
	}
}
