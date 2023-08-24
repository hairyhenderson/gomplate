package celext

import (
	"fmt"
	"testing"

	"github.com/flanksource/gomplate/v3/k8s"
	"github.com/google/cel-go/cel"
	"github.com/stretchr/testify/assert"
)

func panIf(err error) {
	if err != nil {
		panic(err)
	}
}

func executeTemplate(t *testing.T, i int, input string, output any, environment map[string]any) {
	env, err := cel.NewEnv(GetCelEnv(environment)...)
	panIf(err)

	ast, issues := env.Compile(input)
	if issues != nil && issues.Err() != nil {
		panIf(err)
	}

	prg, err := env.Program(ast, cel.Globals(environment))
	panIf(err)

	out, _, err := prg.Eval(environment)
	panIf(err)

	assert.EqualValues(t, output, out.Value(), fmt.Sprintf("Test:%d failed", i+1))
}

func TestCelNamespace(t *testing.T) {
	testData := []struct {
		Input  string
		Output any
	}{
		{Input: `regexp.Replace("flank", "rank", "flanksource")`, Output: "ranksource"},
		{Input: `regexp.Replace("nothing", "rank", "flanksource")`, Output: "flanksource"},
		{Input: `regexp.Replace("", "", "flanksource")`, Output: "flanksource"},
		{Input: `filepath.Join(["/home/flanksource", "projects", "gencel"])`, Output: "/home/flanksource/projects/gencel"},
	}

	for i, td := range testData {
		executeTemplate(t, i, td.Input, td.Output, nil)
	}
}

func TestCelMultipleReturns(t *testing.T) {
	testData := []struct {
		Input   string
		Outputs []any
	}{
		// {Input: `base64.Encode("flanksource")`, Outputs: []any{"Zmxhbmtzb3VyY2U=", nil}},
		// {Input: `base64.Decode("Zmxhbmtzb3VyY2U=")`, Outputs: []any{"flanksource", nil}},
		{Input: `data.JSONArray("[\"name\",\"flanksource\"]")`, Outputs: []any{[]any{"name", "flanksource"}, nil}},
	}

	for i, td := range testData {
		executeTemplate(t, i, td.Input, td.Outputs, nil)
	}
}

func TestCelVariadic(t *testing.T) {
	testData := []struct {
		Input  string
		Output any
	}{
		{Input: `math.Add([1,2,3,4,5])`, Output: int64(15)},
		{Input: `math.Mul([1,2,3,4,5])`, Output: int64(120)},
		{Input: `Slice([1,2,3,4,5])`, Output: []any{int64(1), int64(2), int64(3), int64(4), int64(5)}},
	}

	for i, td := range testData {
		executeTemplate(t, i, td.Input, td.Output, nil)
	}
}

func TestCelSliceReturn(t *testing.T) {
	testData := []struct {
		Input  string
		Output any
	}{
		{Input: `Split("-", "open-source")`, Output: []string{"open", "source"}},
	}

	for i, td := range testData {
		executeTemplate(t, i, td.Input, td.Output, nil)
	}
}

func TestCelK8s(t *testing.T) {
	testData := []struct {
		Input  string
		Output any
	}{
		{Input: `k8s.is_healthy(healthy_obj)`, Output: true},
		{Input: `k8s.is_healthy(unhealthy_obj)`, Output: false},
		{Input: `k8s.health(healthy_obj).status`, Output: "Healthy"},
		{Input: `k8s.health(unhealthy_obj).message`, Output: "Back-off 40s restarting failed container=main pod=my-pod_argocd(63674389-f613-11e8-a057-fe5f49266390)"},
		{Input: `k8s.health(unhealthy_obj).ok`, Output: false},
		{Input: `k8s.health(healthy_obj).message`, Output: ""},
	}

	for i, td := range testData {
		environment := map[string]any{
			"healthy_obj":   k8s.TestHealthy,
			"unhealthy_obj": k8s.TestUnhealthy,
		}
		executeTemplate(t, i, td.Input, td.Output, environment)
	}
}
