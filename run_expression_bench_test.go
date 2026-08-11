package gomplate

// Run:
//   go test -run=^$ -bench='BenchmarkRunExpressionContext|BenchmarkCELEnvExtend|BenchmarkCELProgramEvaluation' -benchmem
//
// Capture a heap profile in the project scratch directory:
//   go test -run=^$ -bench=BenchmarkRunExpressionContext/cache=hit -benchmem \
//     -memprofile .tmp/cel.mem.pprof -memprofilerate=1
//   go tool pprof -alloc_space -top -nodecount=25 .tmp/cel.mem.pprof

import (
	"fmt"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"
)

func benchExprEnv(withNestedConfig bool) map[string]any {
	env := map[string]any{
		"id":           "0192f0a4-1234-7000-8000-aaaaaaaaaaaa",
		"name":         "nginx-7c5ddbdf54-abcde",
		"namespace":    "default",
		"config_type":  "Kubernetes::Pod",
		"config_class": "Pod",
		"tags": map[string]any{
			"cluster":   "production",
			"namespace": "default",
		},
	}
	if !withNestedConfig {
		return env
	}

	containers := make([]any, 0, 3)
	for i := range 3 {
		containers = append(containers, map[string]any{
			"name":  fmt.Sprintf("container-%d", i),
			"image": fmt.Sprintf("registry.example.com/app:%d.2.3", i),
			"ports": []any{map[string]any{"containerPort": 8080 + i, "protocol": "TCP"}},
			"env": []any{
				map[string]any{"name": "LOG_LEVEL", "value": "info"},
				map[string]any{"name": "REGION", "value": "us-east-1"},
			},
			"resources": map[string]any{
				"limits":   map[string]any{"cpu": "500m", "memory": "512Mi"},
				"requests": map[string]any{"cpu": "100m", "memory": "128Mi"},
			},
		})
	}
	env["config"] = map[string]any{
		"apiVersion": "v1",
		"kind":       "Pod",
		"metadata": map[string]any{
			"name":      "nginx-7c5ddbdf54-abcde",
			"namespace": "default",
			"labels": map[string]any{
				"app": "nginx", "team": "platform", "env": "production", "version": "v1.2.3",
			},
			"annotations": map[string]any{
				"prometheus.io/scrape": "true",
				"prometheus.io/port":   "8080",
			},
			"ownerReferences": []any{
				map[string]any{"apiVersion": "apps/v1", "kind": "ReplicaSet", "name": "nginx-7c5ddbdf54"},
			},
		},
		"spec":   map[string]any{"containers": containers, "nodeName": "ip-10-0-1-23"},
		"status": map[string]any{"phase": "Running", "podIP": "10.0.5.12", "hostIP": "10.0.1.23"},
	}
	return env
}

func BenchmarkRunExpressionContext(b *testing.B) {
	const expression = `config_type == "Kubernetes::Pod"`
	for _, withConfig := range []bool{false, true} {
		name := "small"
		if withConfig {
			name = "large"
		}
		b.Run("cache=hit/environment="+name, func(b *testing.B) {
			celExpressionCache.Flush()
			env := benchExprEnv(withConfig)
			template := Template{Expression: expression, CacheKey: "benchmark-cache-hit-" + name}
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

func BenchmarkRunExpressionContextCompile(b *testing.B) {
	const expression = `config_type == "Kubernetes::Pod"`
	for _, withConfig := range []bool{false, true} {
		name := "small"
		if withConfig {
			name = "large"
		}
		b.Run("cache=miss/environment="+name, func(b *testing.B) {
			env := benchExprEnv(withConfig)
			template := Template{Expression: expression, CelEnvs: []cel.EnvOption{benchmarkNoopFunction()}}
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

func BenchmarkCELEnvExtend(b *testing.B) {
	base, err := baseCelEnv()
	if err != nil {
		b.Fatal(err)
	}
	cases := []struct {
		variables int
		functions int
	}{{1, 0}, {10, 0}, {100, 0}}
	for _, benchmark := range cases {
		name := fmt.Sprintf("variables=%d/functions=%d", benchmark.variables, benchmark.functions)
		b.Run(name, func(b *testing.B) {
			options := benchmarkEnvOptions(benchmark.variables, benchmark.functions)
			b.ReportAllocs()
			for b.Loop() {
				if _, err := base.Extend(options...); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func benchmarkNoopFunction() cel.EnvOption {
	return cel.Function("bench_noop", cel.Overload(
		"bench_noop_string", []*cel.Type{cel.StringType}, cel.StringType,
		cel.UnaryBinding(func(value ref.Val) ref.Val { return value }),
	))
}

func benchmarkEnvOptions(variables, functions int) []cel.EnvOption {
	options := make([]cel.EnvOption, 0, variables+functions)
	for i := range variables {
		options = append(options, cel.Variable(fmt.Sprintf("value_%d", i), cel.AnyType))
	}
	if functions == 1 {
		options = append(options, benchmarkNoopFunction())
	}
	return options
}

func assertBenchmarkExpression(b *testing.B, environment map[string]any, template Template) {
	b.Helper()
	output, err := RunExpression(environment, template)
	if err != nil {
		b.Fatal(err)
	}
	if output != true {
		b.Fatalf("unexpected warm-up result: %v", output)
	}
}
