package gomplate

import (
	"time"

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
	BytesAs:      ojg.BytesAsString,
	TimeFormat:   time.RFC3339Nano,
	WriteLimit:   1024,
}

// Serialize iterates over each key-value pair in the input map
// serializes any struct value to map[string]any.
func Serialize(in map[string]any) (map[string]any, error) {
	if in == nil {
		return nil, nil
	}

	// cel supports time.Time and time.Duration natively - save original and then replace it after decomposition
	nativeTypes := make(map[string]any, len(in))
	jp.Walk(in, func(path jp.Expr, value any) {
		switch v := value.(type) {
		case time.Duration, time.Time:
			nativeTypes[path.String()] = v
		}
	})

	out := alt.Alter(in, &opts).(map[string]any)

	for path, v := range nativeTypes {
		expr, err := jp.ParseString(path)
		if err != nil {
			return nil, err
		}

		if err := expr.SetOne(out, v); err != nil {
			return nil, err
		}
	}

	return out, nil
}
