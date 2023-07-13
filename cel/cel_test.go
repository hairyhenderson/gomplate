package cel

import (
	"testing"

	"github.com/flanksource/gomplate/v3/funcs"
	"github.com/google/cel-go/cel"
	"gotest.tools/v3/assert"
)

func panIf(err error) {
	if err != nil {
		panic(err)
	}
}

func executeTemplate(t *testing.T, i int, input string, output any) {
	env, err := cel.NewEnv(funcs.CelEnvOption...)
	panIf(err)

	ast, issues := env.Compile(input)
	if issues != nil && issues.Err() != nil {
		panIf(err)
	}

	prg, err := env.Program(ast)
	panIf(err)

	out, _, err := prg.Eval(map[string]any{})
	panIf(err)

	assert.DeepEqual(t, out.Value(), output)
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
		executeTemplate(t, i, td.Input, td.Output)
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
		executeTemplate(t, i, td.Input, td.Outputs)
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
		executeTemplate(t, i, td.Input, td.Output)
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
		executeTemplate(t, i, td.Input, td.Output)
	}
}
