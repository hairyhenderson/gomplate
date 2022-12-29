package coll

import (
	"github.com/itchyny/gojq"
	"github.com/pkg/errors"
)

// JQ -
func JQ(jqExpr string, in interface{}) (interface{}, error) {
	query, err := gojq.Parse(jqExpr)
	if err != nil {
		return nil, fmt.Errorf("couldn't parse JQ %s: %w", jqExpr, err)
	}
	iter := query.Run(in)
	var out interface{}
	a := []interface{}{}
	for {
		v, ok := iter.Next()
		if !ok {
			break
		}
		if err, ok := v.(error); ok {
			return nil, errors.Wrap(err, "executing JQ failed")
		}
		if v != nil { // TODO: Check, if nil may be a valid result
			a = append(a, v)
		}
	}
	if len(a) == 1 {
		out = a[0]
	} else {
		out = a
	}

	return out, nil
}
