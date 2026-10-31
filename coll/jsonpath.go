package coll

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"k8s.io/client-go/util/jsonpath"
)

type missingKeyCtxKey struct{}

// ContextWithMissingKey returns a context with the missing-key policy set.
func ContextWithMissingKey(ctx context.Context, missingKey string) context.Context {
	return context.WithValue(ctx, missingKeyCtxKey{}, missingKey)
}

// MissingKeyFromContext returns the missing-key policy from the context.
// Defaults to "error" if not set.
func MissingKeyFromContext(ctx context.Context) string {
	if ctx == nil {
		return "error"
	}
	if v, ok := ctx.Value(missingKeyCtxKey{}).(string); ok && v != "" {
		return v
	}
	return "error"
}

// JSONPath -
func JSONPath(p string, in any) (any, error) {
	return JSONPathWithContext(context.Background(), p, in)
}

// JSONPathWithContext -
func JSONPathWithContext(ctx context.Context, p string, in any) (any, error) {
	missingKey := MissingKeyFromContext(ctx)
	allowMissingKeys := missingKey == "zero" || missingKey == "default" || missingKey == "invalid"

	jp, err := parsePath(p, allowMissingKeys)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse JSONPath %s: %w", p, err)
	}
	results, err := jp.FindResults(in)
	if err != nil {
		if allowMissingKeys && strings.Contains(err.Error(), "array index out of bounds") {
			if matches, ok := evalMultiBranchBounds(ctx, p, in); ok {
				return matches, nil
			}
			return nil, nil
		}
		return nil, fmt.Errorf("executing JSONPath failed: %w", err)
	}

	if len(results) == 1 && len(results[0]) == 1 {
		return extractResult(results[0][0])
	}

	a, err := collectResults(results)
	if err != nil {
		return nil, err
	}
	if allowMissingKeys && len(a) == 0 {
		return nil, nil
	}
	return a, nil
}

func evalMultiBranchBounds(ctx context.Context, p string, in any) (any, bool) {
	idx := strings.Index(p, "[*]")
	if idx == -1 {
		return nil, false
	}
	prefix := p[:idx]
	subPath := p[idx+3:]
	if subPath != "" && !strings.HasPrefix(subPath, ".") && !strings.HasPrefix(subPath, "[") {
		subPath = "." + subPath
	}

	var items []any
	if prefix == "" || prefix == "$" || prefix == "." {
		items = toSlice(in)
	} else {
		prefixRes, err := JSONPathWithContext(ctx, prefix, in)
		if err != nil || prefixRes == nil {
			return nil, false
		}
		items = toSlice(prefixRes)
	}

	if items == nil {
		return nil, false
	}

	var matches []any
	for _, item := range items {
		var res any
		var err error
		if subPath == "" {
			res = item
		} else {
			res, err = JSONPathWithContext(ctx, subPath, item)
		}
		if err == nil && res != nil {
			if s, ok := res.([]any); ok {
				matches = append(matches, s...)
			} else {
				matches = append(matches, res)
			}
		}
	}

	if len(matches) == 0 {
		return nil, true
	}
	return matches, true
}

func toSlice(in any) []any {
	if in == nil {
		return nil
	}
	if s, ok := in.([]any); ok {
		return s
	}
	val := reflect.ValueOf(in)
	for val.Kind() == reflect.Pointer || val.Kind() == reflect.Interface {
		if val.IsNil() {
			return nil
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Slice && val.Kind() != reflect.Array {
		return nil
	}
	s := make([]any, val.Len())
	for i := 0; i < val.Len(); i++ {
		s[i] = val.Index(i).Interface()
	}
	return s
}

func collectResults(results [][]reflect.Value) ([]any, error) {
	a := make([]any, 0)
	for _, r := range results {
		for _, v := range r {
			o, err := extractResult(v)
			if err != nil {
				return nil, err
			}
			if o != nil {
				a = append(a, o)
			}
		}
	}
	return a, nil
}

func parsePath(p string, allowMissingKeys bool) (*jsonpath.JSONPath, error) {
	jp := jsonpath.New("<jsonpath>")
	err := jp.Parse("{" + p + "}")
	if err != nil {
		return nil, err
	}
	jp.AllowMissingKeys(allowMissingKeys)
	return jp, nil
}

func extractResult(v reflect.Value) (any, error) {
	// a nil input gives an invalid (zero) value
	if !v.IsValid() {
		return nil, nil
	}

	if v.CanInterface() {
		return v.Interface(), nil
	}

	return nil, fmt.Errorf("JSONPath couldn't access field")
}
