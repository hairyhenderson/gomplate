package strings

import (
	"time"

	"github.com/flanksource/gomplate/v3/conv"
	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// var durationHumanize = cel.Function("HumanDuration",
// 	cel.Overload("duration.HumanDuration",
// 		[]*cel.Type{cel.DynType},
// 		cel.StringType,
// 		cel.UnaryBinding(func(value ref.Val) ref.Val {
// 			return types.String(HumanDuration(value.Value()))
// 		}),
// 	),
// )

func age(value ref.Val) ref.Val {
	switch v := value.Value().(type) {
	case int, int32, int64, float32, float64:
		return types.Duration{Duration: time.Duration(conv.ToInt64(v))}
	case string:
		return types.Duration{Duration: Age(v)}
	}
	return types.NewErr("cannot convert %v into duration", value.Type())
}

var DurationAgeGen = cel.Function("Age",
	cel.Overload("duration.Age",
		[]*cel.Type{cel.AnyType},
		cel.DurationType,
		cel.UnaryBinding(age),
	),
)
var DurationAgeGen2 = cel.Function("age",
	cel.Overload("duration.age",
		[]*cel.Type{cel.AnyType},
		cel.DurationType,
		cel.UnaryBinding(age),
	),
)

var Durations = []cel.EnvOption{
	cel.Function("duration",
		cel.Overload("double.duration",
			[]*cel.Type{cel.DoubleType},
			cel.DurationType,
			cel.UnaryBinding(func(value ref.Val) ref.Val {
				return types.Duration{Duration: time.Duration(conv.ToInt64(value.Value()))}
			}),
		),
	),
	cel.Function("Duration",
		cel.Overload("duration.Duration",
			[]*cel.Type{cel.StringType},
			cel.DurationType,
			cel.UnaryBinding(func(value ref.Val) ref.Val {
				a, err := ParseDuration(value.Value().(string))
				if err != nil || a == nil {
					return types.Duration{}
				}
				return types.Duration{Duration: *a}
			}),
		),
	),
}
