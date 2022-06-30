package binder

import (
	"reflect"
	"strings"
)

// topicBinder maps params in topic to a struct.
type topicBinder struct{}

func (topicBinder) Name() string {
	return "topic"
}

func (topicBinder) Bind(m map[string][]string, obj interface{}) error {
	return iterFields(obj, func(field reflect.StructField, value reflect.Value) error {
		key := field.Tag.Get("topic")
		if levels, ok := m[key]; ok {
			if field.Type == reflect.TypeOf(levels) {
				value.Set(reflect.ValueOf(levels))
			} else if err := setWithProperType(field.Type.Kind(), strings.Join(levels, "/"), value); err != nil {
				return err
			}
		}
		return nil
	})
}
