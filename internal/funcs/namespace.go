package funcs

import (
	"fmt"
	"reflect"
	"strings"
)

var _ fmt.Stringer = (*namespace)(nil)

type namespace struct {
	self any //must be pointer to outer struct
}

func (n *namespace) String() string {
	ns := n.self
	if ns == nil {
		return "<namespace>"
	}

	nsType := reflect.TypeOf(ns)
	if nsType.Kind() != reflect.Pointer || nsType.Elem().Kind() != reflect.Struct {
		panic("invalid namespace type " + nsType.String() + ": must be pointer to struct")
	}

	var public []string
	public = appendPublicMethods(nsType, public)
	nsType = nsType.Elem()
	public = appendPublicFields(nsType, public)

	nsName := nsType.String()
	nsName = strings.TrimPrefix(nsName, "funcs.")
	nsName = strings.TrimSuffix(nsName, "Funcs")

	return fmt.Sprintf("<namespace %s %s>", nsName, public)
}

func appendPublicFields(nsType reflect.Type, public []string) []string {
	for _, field := range reflect.VisibleFields(nsType) {
		if !field.IsExported() {
			continue
		}

		public = append(public, field.Name)
	}

	return public
}

func appendPublicMethods(nsType reflect.Type, public []string) []string {
	for i := range nsType.NumMethod() {
		method := nsType.Method(i)
		if !method.IsExported() {
			continue
		}

		if method.Name == "String" && nsType.Implements(reflect.TypeFor[fmt.Stringer]()) {
			continue
		}

		public = append(public, method.Name)
	}

	return public
}
