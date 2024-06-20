package funcs

import (
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

var collListFirst = cel.Function("first",
	cel.MemberOverload("slice_first_any",
		[]*cel.Type{
			cel.ListType(cel.DynType),
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			list, err := args[0].ConvertToNative(reflect.TypeOf([]any{}))
			if err != nil {
				return types.WrapErr(err)
			}

			l, ok := list.([]any)
			if !ok {
				return types.WrapErr(err)
			}

			if len(l) == 0 {
				return types.DefaultTypeAdapter.NativeToValue("")
			}

			return types.DefaultTypeAdapter.NativeToValue(l[0])
		}),
	),
)

var collListLast = cel.Function("last",
	cel.MemberOverload("slice_last_any",
		[]*cel.Type{
			cel.ListType(cel.DynType),
		},
		cel.DynType,
		cel.FunctionBinding(func(args ...ref.Val) ref.Val {
			list, err := args[0].ConvertToNative(reflect.TypeOf([]any{}))
			if err != nil {
				return types.WrapErr(err)
			}

			l, ok := list.([]any)
			if !ok {
				return types.WrapErr(err)
			}

			if len(l) == 0 {
				return types.DefaultTypeAdapter.NativeToValue("")
			}

			return types.DefaultTypeAdapter.NativeToValue(l[len(l)-1])
		}),
	),
)
