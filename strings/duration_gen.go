package strings

import (
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

var durationAgeGen = cel.Function("Age",
	cel.Overload("duration.Age",
		[]*cel.Type{cel.StringType},
		cel.DurationType,
		cel.UnaryBinding(func(value ref.Val) ref.Val {
			a := Age(value.Value().(string))
			return types.Duration{Duration: a}
		}),
	),
)

var durationDurationGen = cel.Function("Duration",
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
)
