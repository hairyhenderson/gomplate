package gomplate

// This benchmark exercises the CEL expression evaluation path (RunExpressionContext)
// the way callers use it at runtime: with a CacheKey so the compiled cel.Program is
// served from celExpressionCache on every iteration after the first.
//
// Motivation: a production heap profile showed RunExpressionContext -> GetCelEnv
// (notably kubernetes.Library()) + Serialize accounting for the single largest slice
// of lifetime allocation, because GetCelEnv was rebuilt on EVERY evaluation even
// though the compiled program is cached. After iteration 1 (the cache miss), all
// remaining iterations are cache hits; any allocation that remains is the per-call
// overhead that runs regardless of the program cache.
//
// Only the cache-hit steady state is benchmarked: that is the prod hot path and the
// regression guard for the "GetCelEnv must not run on cache hits" invariant.
//
// Run:
//   go test -run=^$ -bench=BenchmarkRunExpressionContext -benchmem
//
// Capture a heap profile and inspect it the same way we inspect prod ones:
//   go test -run=^$ -bench=BenchmarkRunExpressionContext/cacheHit -benchmem \
//     -memprofile /tmp/cel.mem.pprof -memprofilerate=1
//   go tool pprof -alloc_space -top -nodecount=25 /tmp/cel.mem.pprof
//   go tool pprof -alloc_space -peek 'GetCelEnv$' /tmp/cel.mem.pprof

import (
	"fmt"
	"testing"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"
)

// benchExprEnv returns an env map shaped like a Kubernetes Pod config item as the
// scraper passes it to template evaluation. GetCelEnv registers one cel.Variable per
// top-level key and Serialize walks the entire structure, so env size directly drives
// the per-call allocation under test.
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

	if withNestedConfig {
		containers := make([]any, 0, 3)
		for i := 0; i < 3; i++ {
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
	}

	return env
}

// exprBenchSink prevents the compiler from optimizing away results.
var exprBenchSink any

// BenchmarkRunExpressionContext measures the CEL evaluation path on the cache-hit
// steady state: a CacheKey is set so the compiled cel.Program is reused from
// celExpressionCache. After the warm-up run, every iteration is a cache hit, and the
// reported B/op / allocs/op is the per-call overhead that runs regardless of the program
// cache (Serialize + Eval, plus GetCelEnv if a regression reintroduces it before the
// cache lookup).
func BenchmarkRunExpressionContext(b *testing.B) {
	const expression = `config_type == "Kubernetes::Pod"`

	for _, withConfig := range []bool{false, true} {
		name := "cacheHit/smallEnv"
		if withConfig {
			name = "cacheHit/largeEnv"
		}
		b.Run(name, func(b *testing.B) {
			env := benchExprEnv(withConfig)
			tmpl := Template{
				Expression: expression,
				CacheKey:   "bench.RunExpressionContext:config_type==Kubernetes::Pod",
			}
			// Warm the cache once so we measure steady state, not the one-time compile.
			if _, err := RunExpression(env, tmpl); err != nil {
				b.Fatal(err)
			}

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				out, err := RunExpression(env, tmpl)
				if err != nil {
					b.Fatal(err)
				}
				exprBenchSink = out
			}
		})
	}
}

// BenchmarkRunExpressionContextCompile measures the CEL compile path
// (RunExpressionContext cache MISS). Production expressions that reference a
// context-capturing function such as catalog.query attach a CelEnv, which makes
// the compiled program non-cacheable (IsCacheable() is false when len(CelEnvs)
// != 0), so they pay the full env-build + compile cost on EVERY call. This is the
// path that previously rebuilt kubernetes.Library() and revalidated all of its
// declarations every time; it is now served by extending the cached base env.
func BenchmarkRunExpressionContextCompile(b *testing.B) {
	const expression = `config_type == "Kubernetes::Pod"`

	// A trivial CelEnv: its only purpose is to make the template non-cacheable so
	// every iteration goes through the compile path (mirrors catalog.query & co.).
	noopFn := cel.Function("bench_noop",
		cel.Overload("bench_noop_string",
			[]*cel.Type{cel.StringType}, cel.StringType,
			cel.UnaryBinding(func(v ref.Val) ref.Val { return v }),
		),
	)

	for _, withConfig := range []bool{false, true} {
		name := "compile/smallEnv"
		if withConfig {
			name = "compile/largeEnv"
		}
		b.Run(name, func(b *testing.B) {
			env := benchExprEnv(withConfig)

			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				out, err := RunExpression(env, Template{
					Expression: expression,
					CelEnvs:    []cel.EnvOption{noopFn},
				})
				if err != nil {
					b.Fatal(err)
				}
				exprBenchSink = out
			}
		})
	}
}
