package gomplate

import (
	"fmt"
	"testing"

	"github.com/google/cel-go/cel"
)

type benchmarkNativeInput struct {
	DisplayName string `json:"display_name"`
	Scores      []int  `json:"scores"`
}

func BenchmarkCELEnvExtendCustomFunction(b *testing.B) {
	base, err := baseCelEnv()
	if err != nil {
		b.Fatal(err)
	}
	options := benchmarkEnvOptions(10, 1)
	b.ReportAllocs()
	for b.Loop() {
		if _, err := base.Extend(options...); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkCELProgramEvaluation(b *testing.B) {
	cases := []struct {
		name       string
		expression string
		data       map[string]any
	}{
		{"scalar", `config_type == "Kubernetes::Pod"`, map[string]any{"config_type": "Kubernetes::Pod"}},
		{"list_optional_regex", `([1, 2] + [3, 4]).size() == 4 && optional.of(name).orValue("") == "Ada" && name.matches("^A.*")`, map[string]any{"name": "Ada"}},
		{"comprehension", `[1, 2, 3, 4, 5].filter(n, n % 2 == 0).map(n, n * n).exists(n, n == 16)`, nil},
	}
	for _, benchmark := range cases {
		for _, optimize := range []bool{false, true} {
			name := fmt.Sprintf("expression=%s/optimized=%t", benchmark.name, optimize)
			b.Run(name, func(b *testing.B) {
				data, err := serializeForCEL(benchmark.data, currentNativeTypes())
				if err != nil {
					b.Fatal(err)
				}
				program, err := compileBenchmarkCELProgram(data, benchmark.expression, optimize)
				if err != nil {
					b.Fatal(err)
				}
				if output, _, err := program.Eval(data); err != nil || output.Value() != true {
					b.Fatalf("unexpected warm-up result %v: %v", output, err)
				}
				b.ReportAllocs()
				for b.Loop() {
					if _, _, err := program.Eval(data); err != nil {
						b.Fatal(err)
					}
				}
			})
		}
	}
}

func BenchmarkRunExpressionNativeInput(b *testing.B) {
	previousTypes := currentNativeTypes()
	b.Cleanup(func() {
		nativeTypeRegistry.Lock()
		defer nativeTypeRegistry.Unlock()
		nativeTypeRegistry.snapshot.Store(previousTypes)
	})
	if err := RegisterType(benchmarkNativeInput{}); err != nil {
		b.Fatal(err)
	}
	cases := []struct {
		name  string
		value any
	}{
		{"map", map[string]any{"display_name": "Ada", "scores": []int{1, 2, 3}}},
		{"native_struct", benchmarkNativeInput{DisplayName: "Ada", Scores: []int{1, 2, 3}}},
	}
	for _, benchmark := range cases {
		b.Run("input="+benchmark.name, func(b *testing.B) {
			env := map[string]any{"person": benchmark.value}
			template := Template{
				Expression: `person.display_name == "Ada" && person.scores.size() == 3`,
				CacheKey:   "benchmark-native-input-" + benchmark.name,
			}
			assertBenchmarkExpression(b, env, template)
			b.ReportAllocs()
			for b.Loop() {
				if _, err := RunExpression(env, template); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func compileBenchmarkCELProgram(data map[string]any, expression string, optimize bool) (cel.Program, error) {
	base, err := baseCelEnv()
	if err != nil {
		return nil, err
	}
	env, err := base.Extend(celEnvOptions(data, Template{}, currentNativeTypes())...)
	if err != nil {
		return nil, err
	}
	ast, issues := env.Compile(expression)
	if issues != nil && issues.Err() != nil {
		return nil, issues.Err()
	}
	if optimize {
		return env.Program(ast, cel.EvalOptions(cel.OptOptimize))
	}
	return env.Program(ast)
}
