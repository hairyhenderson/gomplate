package funcs

import (
	"fmt"

	"github.com/google/cel-go/common/types/ref"
)

func convertMap(arg ref.Val) (map[string]any, error) {
	switch m := arg.Value().(type) {
	case map[ref.Val]ref.Val:
		var out = make(map[string]any)
		for key, val := range m {
			out[key.Value().(string)] = val.Value()
		}
		return out, nil
	case map[string]any:
		return m, nil
	default:
		return nil, fmt.Errorf("Not a map %T\n", arg.Value())
	}
}

func sliceToNative[K any](args ...ref.Val) ([]K, error) {
	var out []K

	for _, arg := range args {
		list, ok := arg.Value().([]ref.Val)
		if !ok {
			return nil, fmt.Errorf("not a list %T", arg.Value())
		}

		for _, val := range list {
			out = append(out, val.Value().(K))
		}
	}

	return out, nil
}
