package coll

import (
	"fmt"

	"github.com/itchyny/gojq"
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
			return nil, fmt.Errorf("executing JQ failed: %w", err)
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
