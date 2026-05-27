package gomplate

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ohler55/ojg"
	"github.com/ohler55/ojg/alt"
	"github.com/ohler55/ojg/jp"
	"github.com/ohler55/ojg/oj"
)

var opts = oj.Options{
	Color:        false,
	InitSize:     256,
	CreateKey:    "",
	FullTypePath: false,
	OmitNil:      false,
	OmitEmpty:    false,
	UseTags:      true,
	KeyExact:     true,
	NestEmbed:    false,

	BytesAs:    ojg.BytesAsString,
	TimeFormat: "time",
	WriteLimit: 1024,
}

type AsMapper interface {
	AsMap(fields ...string) map[string]any
}

type nativeType struct {
	path jp.Expr
	val  any
}

// Serialize iterates over each key-value pair in the input map
// serializes any struct value to map[string]any.
func Serialize(in map[string]any) (out map[string]any, err error) {
	if in == nil {
		return nil, nil
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic during serialization: %v", r)
		}
	}()

	// cel supports time.Duration natively - save original and then replace it after decomposition
	// FIXME: This does not work for anything inside Structs
	nativeTypes := make([]nativeType, 0, len(in))
	jp.Walk(in, func(path jp.Expr, value any) {
		add := func(v any) {
			// Copy path so later Walk mutations can not affect the stored expression.
			nativeTypes = append(nativeTypes, nativeType{path: append(jp.Expr(nil), path...), val: v})
		}

		switch v := value.(type) {
		case AsMapper:
			add(v.AsMap())
		case uuid.UUID:
			add(v.String())
		case *uuid.UUID:
			if v != nil {
				add(v.String())
			}
		case time.Duration:
			add(v)
		}
	})

	out = alt.Alter(in, &opts).(map[string]any)

	for _, native := range nativeTypes {
		if err := native.path.SetOne(out, native.val); err != nil {
			return nil, err
		}
	}
	return out, nil
}
