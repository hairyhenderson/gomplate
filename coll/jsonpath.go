package coll

import (
	"reflect"

	"github.com/pkg/errors"
	"k8s.io/client-go/util/jsonpath"
)

// JSONPath -
func JSONPath(p string, in interface{}) (interface{}, error) {
	jp, err := parsePath(p)
	if err != nil {
		return nil, errors.Wrapf(err, "couldn't parse JSONPath %s", p)
	}
	results, err := jp.FindResults(in)
	if err != nil {
		return nil, errors.Wrap(err, "executing JSONPath failed")
	}

	var out interface{}
	if len(results) == 1 && len(results[0]) == 1 {
		v := results[0][0]
		out, err = extractResult(v)
		if err != nil {
			return nil, err
		}
	} else {
		a := []interface{}{}
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

func extractResult(v reflect.Value) (interface{}, error) {
	if v.CanInterface() {
		return v.Interface(), nil
	}

	return nil, errors.Errorf("JSONPath couldn't access field")
}
