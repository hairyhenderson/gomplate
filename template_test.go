package gomplate

import (
	"testing"
	"time"

	_ "github.com/flanksource/gomplate/v3/js"
	_ "github.com/robertkrimen/otto/underscore"
)

func TestCacheKeyConsistency(t *testing.T) {
	var hello = func() any {
		return "world"
	}

	var foo = func() any {
		return "bar"
	}

	{
		tt := Template{
			Expression: "{{.name}}{{.age}}",
			Functions: map[string]any{
				"hello": hello,
				"Hello": foo,
				"foo":   foo,
				"Foo":   foo,
			},
		}

		expectedCacheKey := tt.cacheKey(map[string]any{"age": 19, "name": "james"})
		for i := 0; i < 10; i++ {
			key := tt.cacheKey(map[string]any{"age": 19, "name": "james"})
			if key != expectedCacheKey {
				t.Errorf("cache key mismatch: %s != %s", key, expectedCacheKey)
			}
		}
	}

	{
		tt := Template{
			Template:   "{{.name}}{{.age}}",
			LeftDelim:  "{{",
			RightDelim: "}}",
			Functions: map[string]any{
				"hello": hello,
				"Hello": foo,
				"foo":   foo,
				"Foo":   foo,
			},
		}

		expectCacheKey := tt.cacheKey(map[string]any{"age": 19, "name": "james"})
		for i := 0; i < 10; i++ {
			key := tt.cacheKey(map[string]any{"age": 19, "name": "james"})
			if key != expectCacheKey {
				t.Errorf("cache key mismatch: %s != %s", key, expectCacheKey)
			}
		}
	}
}

func TestExplicitCacheKey(t *testing.T) {
	tt := Template{
		Expression: "name + age",
		CacheKey:   "user-defined-key",
	}

	if got := tt.cacheKey(map[string]any{"foo": 1}); got != "user-defined-key" {
		t.Errorf("expected explicit CacheKey to be used, got %q", got)
	}

	if got := tt.cacheKey(map[string]any{"bar": 2}); got != "user-defined-key" {
		t.Errorf("explicit CacheKey must be stable across env shapes, got %q", got)
	}

	if !tt.IsCacheable() {
		t.Errorf("template with explicit CacheKey must be cacheable")
	}

	withFuncs := Template{
		Expression: "name",
		Functions:  map[string]any{"hello": func() any { return "world" }},
	}
	if withFuncs.IsCacheable() {
		t.Errorf("template with Functions and no CacheKey must not be cacheable")
	}
	withFuncs.CacheKey = "stable"
	if !withFuncs.IsCacheable() {
		t.Errorf("template with Functions but explicit CacheKey must be cacheable")
	}
}

func TestCacheTime(t *testing.T) {
	// No expiration: cache entry should have zero expiration time.
	{
		tpl := Template{
			Expression: "1 + 1",
			CacheKey:   "cachetime-noexp",
			CacheTime:  -1,
		}
		if _, err := RunExpression(nil, tpl); err != nil {
			t.Fatalf("eval: %v", err)
		}
		_, exp, ok := celExpressionCache.GetWithExpiration(tpl.celCacheKey(nil, currentNativeTypes().generation))
		if !ok {
			t.Fatalf("entry not cached")
		}
		if !exp.IsZero() {
			t.Errorf("expected no-expiration entry, got expiry %v", exp)
		}
	}

	// Explicit short TTL: entry expiration should be close to now+CacheTime.
	{
		tpl := Template{
			Expression: "1 + 1",
			CacheKey:   "cachetime-short",
			CacheTime:  50 * time.Millisecond,
		}
		before := time.Now()
		if _, err := RunExpression(nil, tpl); err != nil {
			t.Fatalf("eval: %v", err)
		}
		_, exp, ok := celExpressionCache.GetWithExpiration(tpl.celCacheKey(nil, currentNativeTypes().generation))
		if !ok {
			t.Fatalf("entry not cached")
		}
		diff := exp.Sub(before)
		if diff < 40*time.Millisecond || diff > 200*time.Millisecond {
			t.Errorf("expected expiry ~50ms after set, got %v", diff)
		}
	}

	// Zero CacheTime: should fall back to the cache's default TTL (~1h).
	{
		tpl := Template{
			Expression: "1 + 1",
			CacheKey:   "cachetime-default",
		}
		before := time.Now()
		if _, err := RunExpression(nil, tpl); err != nil {
			t.Fatalf("eval: %v", err)
		}
		_, exp, ok := celExpressionCache.GetWithExpiration(tpl.celCacheKey(nil, currentNativeTypes().generation))
		if !ok {
			t.Fatalf("entry not cached")
		}
		diff := exp.Sub(before)
		if diff < 30*time.Minute || diff > 90*time.Minute {
			t.Errorf("expected ~1h default TTL, got %v", diff)
		}
	}

}

func TestRunTemplate_ValueFunctions(t *testing.T) {
	out, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template:       "hello {{ msg }}",
		ValueFunctions: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello world" {
		t.Errorf("got %q, want %q", out, "hello world")
	}
}

func TestRunTemplate_ValueFunctionsCoexistWithDotAccess(t *testing.T) {
	out, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template:       "{{ msg }}-{{ .msg }}",
		ValueFunctions: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "world-world" {
		t.Errorf("got %q, want %q", out, "world-world")
	}
}

func TestRunTemplate_ValueFunctionsOffKeepsDotOnly(t *testing.T) {
	out, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template: "{{ .msg }}",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "world" {
		t.Errorf("got %q, want %q", out, "world")
	}

	if _, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template: "{{ msg }}",
	}); err == nil {
		t.Error("expected error calling bare {{ msg }} without ValueFunctions, got nil")
	}
}

func TestRunTemplate_ValueFunctionsDoNotMutateCallerFuncs(t *testing.T) {
	callerFuncs := map[string]any{}
	_, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template:       "{{ msg }}",
		Functions:      callerFuncs,
		ValueFunctions: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(callerFuncs) != 0 {
		t.Errorf("caller Functions map was mutated: %v", callerFuncs)
	}
}

func TestRunTemplate_DelimSetsMultiPass(t *testing.T) {
	out, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template: "hello $(msg)",
		DelimSets: []Delims{
			{Left: "{{", Right: "}}"},
			{Left: "$(", Right: ")"},
		},
		ValueFunctions: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "hello world" {
		t.Errorf("got %q, want %q", out, "hello world")
	}
}

func TestRunTemplate_DelimSetsFeedsOutputForward(t *testing.T) {
	// First pass replaces $(inner) with "msg", producing "{{ msg }}";
	// second pass then resolves the {{ msg }}.
	out, err := RunTemplate(
		map[string]any{"inner": "msg", "msg": "world"},
		Template{
			Template: "{{ $(inner) }}",
			DelimSets: []Delims{
				{Left: "$(", Right: ")"},
				{Left: "{{", Right: "}}"},
			},
			ValueFunctions: true,
		},
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "world" {
		t.Errorf("got %q, want %q", out, "world")
	}
}

func TestRunTemplate_HeaderOverridesDelimSets(t *testing.T) {
	// Header sets [[ ]] so DelimSets should be ignored and $(...) left literal.
	out, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template: "# gotemplate: left-delim=[[ right-delim=]]\n[[ .msg ]] $(msg)",
		DelimSets: []Delims{
			{Left: "{{", Right: "}}"},
			{Left: "$(", Right: ")"},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "world $(msg)" {
		t.Errorf("got %q, want %q", out, "world $(msg)")
	}
}

func TestRunTemplate_DelimSetsPreferredOverSingleDelimPair(t *testing.T) {
	// Both DelimSets and LeftDelim/RightDelim set — DelimSets wins.
	out, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template:   "$(msg)",
		LeftDelim:  "[[",
		RightDelim: "]]",
		DelimSets: []Delims{
			{Left: "$(", Right: ")"},
		},
		ValueFunctions: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "world" {
		t.Errorf("got %q, want %q", out, "world")
	}
}

func TestRunTemplate_SingleDelimPairStillWorks(t *testing.T) {
	out, err := RunTemplate(map[string]any{"msg": "world"}, Template{
		Template:   "[[ .msg ]]",
		LeftDelim:  "[[",
		RightDelim: "]]",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "world" {
		t.Errorf("got %q, want %q", out, "world")
	}
}

func TestRunExpressionReusesProgramAcrossDifferentData(t *testing.T) {
	tpl := Template{
		Expression: "name + age",
		CacheKey:   "reuse-test",
	}

	out1, err := RunExpression(map[string]any{"name": "alice-", "age": "30"}, tpl)
	if err != nil {
		t.Fatalf("first eval: %v", err)
	}
	if out1 != "alice-30" {
		t.Errorf("first eval result: got %v", out1)
	}

	out2, err := RunExpression(map[string]any{"name": "bob-", "age": "42"}, tpl)
	if err != nil {
		t.Fatalf("second eval: %v", err)
	}
	if out2 != "bob-42" {
		t.Errorf("second eval result: got %v (cached program should not leak first-call data)", out2)
	}
}
