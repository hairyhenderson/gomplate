package gomplate

import (
	"testing"

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

		expectedCacheKey := tt.CacheKey(map[string]any{"age": 19, "name": "james"})
		for i := 0; i < 10; i++ {
			key := tt.CacheKey(map[string]any{"age": 19, "name": "james"})
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

		expectCacheKey := tt.CacheKey(map[string]any{"age": 19, "name": "james"})
		for i := 0; i < 10; i++ {
			key := tt.CacheKey(map[string]any{"age": 19, "name": "james"})
			if key != expectCacheKey {
				t.Errorf("cache key mismatch: %s != %s", key, expectCacheKey)
			}
		}
	}
}
