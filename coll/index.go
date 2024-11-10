package coll

import (
	"fmt"
	"math"
	"reflect"
)

// much of the code here is taken from the Go source code, in particular from
// text/template/exec.go and text/template/funcs.go

// Index returns the result of indexing the given map, slice, or array by the
// given index arguments. This is similar to the `index` template function, but
// will return an error if the key is not found. Note that the argument order is
// different from the template function definition found in `funcs/coll.go` to
// allow for variadic indexes.
func Index(v interface{}, keys ...interface{}) (interface{}, error) {
	item := reflect.ValueOf(v)
	item = indirectInterface(item)
	if !item.IsValid() {
		return nil, fmt.Errorf("index of untyped nil")
	}

	indexes := make([]reflect.Value, len(keys))
	for i, k := range keys {
		indexes[i] = reflect.ValueOf(k)
	}

	for _, index := range indexes {
		index = indirectInterface(index)
		var isNil bool
		if item, isNil = indirect(item); isNil {
			return nil, fmt.Errorf("index of nil pointer")
		}
		switch item.Kind() {
		case reflect.Array, reflect.Slice, reflect.String:
			x, err := indexArg(index, item.Len())
			if err != nil {
				return nil, err
			}

			item = item.Index(x)
		case reflect.Map:
			x, err := prepareArg(index, item.Type().Key())
			if err != nil {
				return nil, err
			}

			if v := item.MapIndex(x); v.IsValid() {
				item = v
			} else {
				// the map doesn't contain the key, so return an error
				return nil, fmt.Errorf("map has no key %v", x.Interface())
			}
		case reflect.Invalid:
			// the loop holds invariant: item.IsValid()
			panic("unreachable")
		default:
			return nil, fmt.Errorf("can't index item of type %s", item.Type())
		}
	}

	return item.Interface(), nil
}

// indexArg checks if a reflect.Value can be used as an index, and converts it to int if possible.
//
//nolint:revive
func indexArg(index reflect.Value, cap int) (int, error) {
	var x int64
	switch index.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		x = index.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		val := index.Uint()
		if val > math.MaxInt64 {
			return -1, fmt.Errorf("cannot index slice/array with %d (too large)", val)
		}

		x = int64(val)
	case reflect.Invalid:
		return 0, fmt.Errorf("cannot index slice/array with nil")
	default:
		return 0, fmt.Errorf("cannot index slice/array with type %s", index.Type())
	}

	// note - this has been modified from the original to check for x == cap as
	// well. IMO the original (> only) is a bug.
	if x < 0 || int(x) < 0 || int(x) >= cap {
		return 0, fmt.Errorf("index out of range: %d", x)
	}

	return int(x), nil
}

// prepareArg checks if value can be used as an argument of type argType, and
// converts an invalid value to appropriate zero if possible.
func prepareArg(value reflect.Value, argType reflect.Type) (reflect.Value, error) {
	if !value.IsValid() {
		if !canBeNil(argType) {
			return reflect.Value{}, fmt.Errorf("value is nil; should be of type %s", argType)
		}

		value = reflect.Zero(argType)
	}

	if value.Type().AssignableTo(argType) {
		return value, nil
	}

	if intLike(value.Kind()) && intLike(argType.Kind()) && value.Type().ConvertibleTo(argType) {
		value = value.Convert(argType)

		return value, nil
	}

	return reflect.Value{}, fmt.Errorf("value has type %s; should be %s", value.Type(), argType)
}

var reflectValueType = reflect.TypeOf((*reflect.Value)(nil)).Elem()

// canBeNil reports whether an untyped nil can be assigned to the type. See reflect.Zero.
func canBeNil(typ reflect.Type) bool {
	switch typ.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return true
	case reflect.Struct:
		return typ == reflectValueType
	}

	return false
}

func intLike(typ reflect.Kind) bool {
	switch typ {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return true
	}
	return false
}

// indirect returns the item at the end of indirection, and a bool to indicate
// if it's nil. If the returned bool is true, the returned value's kind will be
// either a pointer or interface.
func indirect(v reflect.Value) (rv reflect.Value, isNil bool) {
	for ; v.Kind() == reflect.Pointer || v.Kind() == reflect.Interface; v = v.Elem() {
		if v.IsNil() {
			return v, true
		}
	}
	return v, false
}

// indirectInterface returns the concrete value in an interface value,
// or else the zero reflect.Value.
// That is, if v represents the interface value x, the result is the same as reflect.ValueOf(x):
// the fact that x was an interface value is forgotten.
func indirectInterface(v reflect.Value) reflect.Value {
	if v.Kind() != reflect.Interface {
		return v
	}
	if v.IsNil() {
		return reflect.Value{}
	}
	return v.Elem()
}
