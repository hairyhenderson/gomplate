package conv

import (
	"fmt"
	"reflect"
)

// InterfaceSlice converts an array or slice of any type into an []any
// for use in functions that expect this.
func InterfaceSlice(slice any) ([]any, error) {
	// avoid all this nonsense if this is already a []any...
	if s, ok := slice.([]any); ok {
		return s, nil
	}
	s := reflect.ValueOf(slice)
	kind := s.Kind()
	switch kind {
	case reflect.Slice, reflect.Array:
		l := s.Len()
		ret := make([]any, l)
		for i := range l {
			ret[i] = s.Index(i).Interface()
		}
		return ret, nil
	default:
		return nil, fmt.Errorf("expected an array or slice, but got a %T", s)
	}
}
