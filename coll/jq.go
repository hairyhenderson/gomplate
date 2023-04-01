package coll

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/itchyny/gojq"
)

// JQ -
func JQ(ctx context.Context, jqExpr string, in interface{}) (interface{}, error) {
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
	var out interface{}
	a := []interface{}{}
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

// jqConvertType converts the input to a map[string]interface{}, []interface{},
// or other supported primitive JSON types.
func jqConvertType(in interface{}) (interface{}, error) {
	// if it's already a supported type, pass it through
	switch in.(type) {
	case map[string]interface{}, []interface{},
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
	if inType.Kind() == reflect.Ptr {
		inType = inType.Elem()
		value = value.Elem()
	}

	mapType := reflect.TypeOf(map[string]interface{}{})
	sliceType := reflect.TypeOf([]interface{}{})
	// if it can be converted to a map or slice, do that
	if inType.ConvertibleTo(mapType) {
		return value.Convert(mapType).Interface(), nil
	} else if inType.ConvertibleTo(sliceType) {
		return value.Convert(sliceType).Interface(), nil
	}

	// if it's a struct, the simplest (though not necessarily most efficient)
	// is to JSON marshal/unmarshal it
	if inType.Kind() == reflect.Struct {
		b, err := json.Marshal(in)
		if err != nil {
			return nil, fmt.Errorf("json marshal struct: %w", err)
		}
		var m map[string]interface{}
		err = json.Unmarshal(b, &m)
		if err != nil {
			return nil, fmt.Errorf("json unmarshal struct: %w", err)
		}
		return m, nil
	}

	// we maybe don't need to convert the value, so return it as-is
	return in, nil
}
