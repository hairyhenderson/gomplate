package tests

import (
	"testing"

	"github.com/flanksource/gomplate/v3"
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
		{structEnv, `results.name + " " + results.Address.city_name`, "Aditya Kathmandu"},
	}

	for _, tc := range tests {
		t.Run(tc.js, func(t *testing.T) {
			out, err := gomplate.RunTemplate(tc.env, gomplate.Template{
				Javascript: tc.js,
			})
			assert.ErrorIs(t, err, nil)
			assert.Equal(t, tc.out, out)
		})
	}
}
