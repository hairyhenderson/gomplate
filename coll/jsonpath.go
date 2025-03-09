package coll

import (
	"fmt"
	"reflect"

	"k8s.io/client-go/util/jsonpath"
)

// JSONPath -
func JSONPath(p string, in any) (any, error) {
	jp, err := parsePath(p)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse JSONPath %s: %w", p, err)
	}
	results, err := jp.FindResults(in)
	if err != nil {
		return nil, fmt.Errorf("executing JSONPath failed: %w", err)
	}

	var out any
	if len(results) == 1 && len(results[0]) == 1 {
		v := results[0][0]
		out, err = extractResult(v)
		if err != nil {
			return nil, err
		}
	} else {
		a := []any{}
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
		out = a
	}

	return out, nil
}

func parsePath(p string) (*jsonpath.JSONPath, error) {
	jp := jsonpath.New("<jsonpath>")
	err := jp.Parse("{" + p + "}")
	if err != nil {
		return nil, err
	}
	jp.AllowMissingKeys(false)
	return jp, nil
}

func extractResult(v reflect.Value) (any, error) {
	if v.CanInterface() {
		return v.Interface(), nil
	}

	return nil, fmt.Errorf("JSONPath couldn't access field")
}
