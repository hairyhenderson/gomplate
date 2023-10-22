package tests

import (
	"testing"
	"time"

	"github.com/flanksource/gomplate/v3"
	_ "github.com/flanksource/gomplate/v3/js"
	"github.com/flanksource/gomplate/v3/kubernetes"
	_ "github.com/robertkrimen/otto/underscore"
	"github.com/stretchr/testify/assert"
)

func TestGomplate(t *testing.T) {
	tests := []struct {
		env      map[string]interface{}
		template string
		out      string
	}{
		{map[string]interface{}{"hello": "world"}, "{{ .hello }}", "world"},
		{map[string]interface{}{"hello": "hello world ?"}, "{{ .hello | urlencode }}", `hello+world+%3F`},
		{map[string]interface{}{"hello": "hello+world+%3F"}, "{{ .hello | urldecode }}", `hello world ?`},
		{map[string]interface{}{"age": 75 * time.Second}, "{{ .age | humanDuration  }}", "1m15s"},
		{map[string]interface{}{"healthySvc": kubernetes.GetUnstructured(kubernetes.TestHealthy)}, "{{ (.healthySvc | isHealthy) }}", "true"},
		{map[string]interface{}{"healthySvc": kubernetes.GetUnstructured(kubernetes.TestLuaStatus)}, "{{ (.healthySvc | getStatus) }}", "Degraded: found less than two generators, Merge requires two or more"},
		{map[string]interface{}{"healthySvc": kubernetes.GetUnstructured(kubernetes.TestHealthy)}, "{{ (.healthySvc | getHealth).Status  }}", "Healthy"},
		{map[string]interface{}{"size": 123456}, "{{ .size | humanSize }}", "120.6K"},
		{map[string]interface{}{"v": "1.2.3-beta.1+c0ff33"}, "{{  (.v | semver).Prerelease  }}", "beta.1"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.3"}, "{{  .old | semverCompare .new }}", "true"},
		{map[string]interface{}{"old": "1.2.3", "new": "1.2.4"}, "{{  .old | semverCompare .new }}", "false"},
		{structEnv, `{{.results.name}} {{.results.Address.city_name}}`, "Aditya Kathmandu"},
		{
			map[string]any{"results": junitEnv},
			`{{.results.passed}}{{ range $r := .results.suites}}{{$r.name}} ✅ {{$r.passed}} ❌ {{$r.failed}} in 🕑 {{$r.duration}}{{end}}`,
			"1hi ✅ 0 ❌ 2 in 🕑 0",
		},
		{
			map[string]any{
				"results": SQLDetails{
					Rows: []map[string]any{{"name": "apm-hub"}, {"name": "config-db"}},
				},
			},
			`{{range $r := .results.rows }}{{range $x, $y := $r }}{{ $y }}{{end}}{{end}}`, "apm-hubconfig-db"},
	}

	for _, tc := range tests {
		t.Run(tc.template, func(t *testing.T) {
			out, err := gomplate.RunTemplate(tc.env, gomplate.Template{
				Template: tc.template,
			})
			assert.ErrorIs(t, err, nil)
			assert.Equal(t, tc.out, out)
		})
	}
}
