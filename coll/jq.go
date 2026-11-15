package coll

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"strconv"

	"github.com/itchyny/gojq"
)

// JQ -
func JQ(ctx context.Context, jqExpr string, in any) (any, error) {
	query, err := gojq.Parse(jqExpr)
	if err != nil {
		return nil, fmt.Errorf("jq parsing expression %q: %w", jqExpr, err)
	}

	// convert input to a supported type, if necessary
	in, err = jqConvertType(in)
	if err != nil {
		return nil, fmt.Errorf("jq type conversion: %w", err)
	}

	iter := query.RunWithContext(ctx, in)
	var out any
	a := []any{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, fmt.Errorf("jq execution: %w", err)
		}
		a = append(a, v)
	}
	if len(a) == 1 {
		out = a[0]
	} else {
		out = a
	}

	return out, nil
}

// jqConvertType converts the input to a map[string]any, []any,
// or other supported primitive JSON types.
func jqConvertType(in any) (any, error) {
	// if it's already a supported type, pass it through
	switch in.(type) {
	case map[string]any, []any,
		string, []byte,
		nil, bool,
		int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return in, nil
	}

	inType := reflect.TypeOf(in)
	value := reflect.ValueOf(in)

	// pointers need to be dereferenced first
	if inType.Kind() == reflect.Pointer {
		inType = inType.Elem()
		value = value.Elem()
	}

	mapType := reflect.TypeOf(map[string]any{})
	sliceType := reflect.TypeOf([]any{})
	// if it can be converted to a map or slice, do that
	if inType.ConvertibleTo(mapType) {
		return value.Convert(mapType).Interface(), nil
	} else if inType.ConvertibleTo(sliceType) {
		return value.Convert(sliceType).Interface(), nil
	}

	// if it's a struct, the simplest (though not necessarily most efficient)
	// is to JSON marshal/unmarshal it
	if inType.Kind() == reflect.Struct {
		return jsonRoundTrip(in)
	}

	// other maps and slices, like map[string]string (.Env) or [][]string (CSV
	// datasources), also need to become map[string]any or []any. Byte slices
	// are passed through as strings.
	switch inType.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		if inType.Elem().Kind() == reflect.Uint8 {
			return in, nil
		}

		return jsonRoundTrip(in)
	}

	// we maybe don't need to convert the value, so return it as-is
	return in, nil
}

// jsonRoundTrip converts in to the generic types gojq supports by marshalling
// it to JSON and back. Numbers are decoded exactly: integers become an int,
// or a float64 if they don't fit in an int but a float64 holds them exactly,
// or otherwise a *big.Int, so large integers aren't rounded.
func jsonRoundTrip(in any) (any, error) {
	b, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("json marshal %T: %w", in, err)
	}

	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()

	var v any
	err = dec.Decode(&v)
	if err != nil {
		return nil, fmt.Errorf("json unmarshal %T: %w", in, err)
	}

	return normalizeJSONNumbers(v), nil
}

// normalizeJSONNumbers replaces each json.Number in v with an int, float64,
// or *big.Int.
func normalizeJSONNumbers(v any) any {
	switch v := v.(type) {
	case map[string]any:
		for k, e := range v {
			v[k] = normalizeJSONNumbers(e)
		}
		return v
	case []any:
		for i, e := range v {
			v[i] = normalizeJSONNumbers(e)
		}
		return v
	case json.Number:
		if i, err := strconv.Atoi(v.String()); err == nil {
			return i
		}
		if i, ok := new(big.Int).SetString(v.String(), 10); ok {
			// prefer a float64 when it holds the value exactly (e.g. 1e20, or
			// anything over 2^31 on 32-bit platforms), so that only integers a
			// float64 would round become a *big.Int
			if f, acc := new(big.Float).SetInt(i).Float64(); acc == big.Exact {
				return f
			}
			return i
		}
		f, _ := v.Float64()
		return f
	default:
		return v
	}
}
