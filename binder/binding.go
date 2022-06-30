package binder

import (
	"github.com/go-playground/validator/v10"
	"reflect"
	"strconv"
)

var validate = validator.New()

// DataBinder describes the interface which needs to
// be implemented for binding the data present in
// the MQTT payload.
type DataBinder interface {
	// Name of the binder
	Name() string
	// Bind unmarshal payload to an object
	Bind([]byte, interface{}) error
}

// Available data binders.
var (
	JSON  = jsonBinder{}
	Text  = textBinder{}
	Topic = topicBinder{}
)

// Validate validates the given struct.
func Validate(obj interface{}) error {
	return validate.Struct(obj)
}

// setWithProperType sets the value in a struct of an indeterminate type to the
// matching value from the request in the same type, so that not all deserialized
// values have to be strings.
// Supported types are string, int, float, and bool.
func setWithProperType(valueKind reflect.Kind, val string, structField reflect.Value) error {
	if valueKind == reflect.Ptr {
		valueKind = structField.Type().Elem().Kind()
		if structField.IsNil() {
			structField.Set(reflect.New(structField.Type().Elem()))
		}
		structField = structField.Elem()
	}
	switch valueKind {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if val == "" {
			val = "0"
		}
		if intVal, err := strconv.ParseInt(val, 10, 64); err != nil {
			return err
		} else {
			structField.SetInt(intVal)
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if val == "" {
			val = "0"
		}
		if uintVal, err := strconv.ParseUint(val, 10, 64); err != nil {
			return err
		} else {
			structField.SetUint(uintVal)
		}
	case reflect.Bool:
		if val == "" {
			val = "false"
		}
		if boolVal, err := strconv.ParseBool(val); err != nil {
			return err
		} else if boolVal {
			structField.SetBool(true)
		}
	case reflect.Float32:
		if val == "" {
			val = "0.0"
		}
		if floatVal, err := strconv.ParseFloat(val, 32); err != nil {
			return err
		} else {
			structField.SetFloat(floatVal)
		}
	case reflect.Float64:
		if val == "" {
			val = "0.0"
		}
		if floatVal, err := strconv.ParseFloat(val, 64); err != nil {
			return err
		} else {
			structField.SetFloat(floatVal)
		}
	case reflect.String:
		structField.SetString(val)
	}
	return nil
}

// iterFields iters each field of a struct and calls the given function with the field.
// When an error is returned, it stops the iteration.
func iterFields(obj interface{}, f func(reflect.StructField, reflect.Value) error) error {
	t := reflect.TypeOf(obj)
	v := reflect.ValueOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		var err error
		if field.Anonymous {
			anonymousType := field.Type
			anonymousValue := v.Field(i)
			if anonymousType.Kind() == reflect.Ptr {
				anonymousType = field.Type.Elem()
			} else {
				anonymousValue = anonymousValue.Addr()
			}
			if anonymousType.Kind() == reflect.Struct {
				if anonymousValue.IsNil() {
					anonymousValue = reflect.New(anonymousType)
					v.Field(i).Set(anonymousValue)
				}
				err = iterFields(anonymousValue.Interface(), f)
			}
		} else {
			err = f(field, v.Field(i))
		}
		if err != nil {
			return err
		}
	}
	return nil
}
