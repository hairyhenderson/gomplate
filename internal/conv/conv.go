package conv

import (
	"fmt"
	"reflect"
)

// InterfaceSlice converts an array or slice of any type into an []interface{}
// for use in functions that expect this.
func InterfaceSlice(slice interface{}) ([]interface{}, error) {
	// avoid all this nonsense if this is already a []interface{}...
	if s, ok := slice.([]interface{}); ok {
		return s, nil
	}
	s := reflect.ValueOf(slice)
	kind := s.Kind()
	switch kind {
	case reflect.Slice, reflect.Array:
		l := s.Len()
		ret := make([]interface{}, l)
		for i := 0; i < l; i++ {
			ret[i] = s.Index(i).Interface()
		}
		return ret, nil
	default:
		return nil, fmt.Errorf("expected an array or slice, but got a %T", s)
	}
}
