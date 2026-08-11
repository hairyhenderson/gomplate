package gomplate

import (
	"fmt"
	"maps"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/ext"
	"google.golang.org/protobuf/proto"
)

type nativeTypeSnapshot struct {
	items        []any
	reflectTypes map[reflect.Type]struct{}
	keys         map[string]struct{}
	generation   uint64
	envOption    cel.EnvOption
}

var emptyNativeTypeSnapshot = &nativeTypeSnapshot{
	reflectTypes: map[reflect.Type]struct{}{},
	keys:         map[string]struct{}{},
}

var nativeTypeRegistry struct {
	sync.Mutex
	snapshot atomic.Pointer[nativeTypeSnapshot]
}

func currentNativeTypes() *nativeTypeSnapshot {
	if snapshot := nativeTypeRegistry.snapshot.Load(); snapshot != nil {
		return snapshot
	}
	return emptyNativeTypeSnapshot
}

// RegisterType makes a Go or CEL type available to subsequent CEL evaluations.
func RegisterType(value any) error {
	item, reflectType, key, err := nativeTypeRegistration(value)
	if err != nil {
		return err
	}

	nativeTypeRegistry.Lock()
	defer nativeTypeRegistry.Unlock()

	current := currentNativeTypes()
	if _, found := current.keys[key]; found {
		return nil
	}

	items := append(append([]any(nil), current.items...), item)
	args := make([]any, 0, len(items)+1)
	args = append(args, ext.ParseStructField(jsonCELFieldName))
	args = append(args, items...)
	envOption := ext.NativeTypes(args...)
	if _, err := cel.NewEnv(envOption); err != nil {
		return fmt.Errorf("register CEL type %s: %w", key, err)
	}

	reflectTypes := maps.Clone(current.reflectTypes)
	addRegisteredReflectType(reflectTypes, reflectType)
	keys := maps.Clone(current.keys)
	keys[key] = struct{}{}
	nativeTypeRegistry.snapshot.Store(&nativeTypeSnapshot{
		items:        items,
		reflectTypes: reflectTypes,
		keys:         keys,
		generation:   current.generation + 1,
		envOption:    envOption,
	})
	return nil
}

func nativeTypeRegistration(value any) (item any, reflectType reflect.Type, key string, err error) {
	if value == nil {
		return nil, nil, "", fmt.Errorf("register CEL type: value is nil")
	}

	switch typed := value.(type) {
	case reflect.Type:
		if typed == nil {
			return nil, nil, "", fmt.Errorf("register CEL type: reflect.Type is nil")
		}
		return typed, typed, registeredReflectTypeKey(typed), nil
	case reflect.Value:
		if !typed.IsValid() {
			return nil, nil, "", fmt.Errorf("register CEL type: reflect.Value is invalid")
		}
		return typed, typed.Type(), registeredReflectTypeKey(typed.Type()), nil
	case proto.Message:
		reflectType = reflect.TypeOf(typed)
		if reflectType.Kind() == reflect.Pointer && reflect.ValueOf(typed).IsNil() {
			return nil, nil, "", fmt.Errorf("register CEL type: protobuf message is nil")
		}
		return typed, reflectType, "proto:" + string(typed.ProtoReflect().Descriptor().FullName()), nil
	case types.StructTypeDescriptor:
		refType, ok := typed.(ref.Type)
		if !ok {
			return nil, nil, "", fmt.Errorf("register CEL type: descriptor %T must also implement ref.Type", typed)
		}
		return refType, typed.ReflectType(), "cel:" + refType.TypeName(), nil
	case ref.Type:
		return typed, nil, "cel:" + typed.TypeName(), nil
	default:
		reflectType = reflect.TypeOf(value)
		return reflectType, reflectType, registeredReflectTypeKey(reflectType), nil
	}
}

func registeredReflectTypeKey(reflectType reflect.Type) string {
	for reflectType.Kind() == reflect.Pointer {
		reflectType = reflectType.Elem()
	}
	return "reflect:" + reflectType.PkgPath() + ":" + reflectType.String()
}

func addRegisteredReflectType(registered map[reflect.Type]struct{}, reflectType reflect.Type) {
	if reflectType == nil {
		return
	}
	registered[reflectType] = struct{}{}
	if reflectType.Kind() == reflect.Pointer {
		registered[reflectType.Elem()] = struct{}{}
	} else {
		registered[reflect.PointerTo(reflectType)] = struct{}{}
	}
}

func (snapshot *nativeTypeSnapshot) preserves(value any) bool {
	if snapshot == nil || value == nil {
		return false
	}
	_, found := snapshot.reflectTypes[reflect.TypeOf(value)]
	return found
}

func jsonCELFieldName(field reflect.StructField) string {
	tag, found := field.Tag.Lookup("json")
	if !found {
		return field.Name
	}
	name := strings.Split(tag, ",")[0]
	if name == "" {
		return field.Name
	}
	return name
}
