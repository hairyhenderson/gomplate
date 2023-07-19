package gomplate

import (
	"testing"
	"time"

	"github.com/flanksource/gomplate/v3/k8s"
	"github.com/stretchr/testify/assert"
)

func TestJavascript(t *testing.T) {
	tests := []struct {
		env map[string]interface{}
		js  string
		out string
	}{
		{map[string]interface{}{"x": 5, "y": 3}, "x + y", "8"},
		{map[string]interface{}{"str": "Hello, World!"}, "str", "Hello, World!"},
		{map[string]interface{}{"numbers": []int{1, 2, 3, 4, 5}}, "_.reduce(numbers, function(memo, num){ return memo + num; }, 0)", "15"},
		{map[string]interface{}{"numbers": []int{4, 2, 55}}, `_.max(numbers)`, "55"},
		{map[string]interface{}{"arr": []int{1, 2, 1, 4, 1, 2}}, "_.uniq(arr)", "1,2,4"},
		{map[string]interface{}{"numbers": []int{4, 2, 55}}, `_.max(numbers)`, "55"},
		{map[string]interface{}{"arr": []int{1, 2, 1, 4, 1, 2}}, "_.uniq(arr)", "1,2,4"},
		{map[string]interface{}{"x": "1Ki"}, "fromSI(x)", "1024"},
		{map[string]interface{}{"x": "2m"}, "fromMillicores(x)", "2"},
		{map[string]interface{}{"name": "mission.compute.internal"}, "k8s.getNodeName(name)", "mission"},
		{map[string]interface{}{"msg": map[string]any{"healthStatus": map[string]string{"status": "HEALTHY"}}}, "k8s.conditions.isReady(msg)", "true"},
	}

	for _, tc := range tests {
		t.Run(tc.js, func(t *testing.T) {
			out, err := RunTemplate(tc.env, Template{
				Javascript: tc.js,
			})
			assert.ErrorIs(t, err, nil)
			assert.Equal(t, tc.out, out)
		})
	}
}

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
