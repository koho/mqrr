package binder

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// textBinder slices a string into all substrings separated by given separator,
// and takes the specific substring according to the given index.
type textBinder struct{}

func (textBinder) Name() string {
	return "text"
}

func (textBinder) Bind(raw []byte, obj interface{}) error {
	if raw == nil {
		return fmt.Errorf("invalid payload")
	}
	return iterFields(obj, func(field reflect.StructField, value reflect.Value) error {
		tag := field.Tag.Get("text")
		if tag == "" {
			return nil
		}
		sep := ""
		parts := strings.SplitN(tag, ",", 2)
		if len(parts) > 1 {
			sep = parts[1]
		}
		idx, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			return err
		}
		text := string(raw)
		var textSplits = []string{text}
		if sep != "" {
			textSplits = strings.Split(text, sep)
		}
		if i := int(idx); i >= 0 && i < len(textSplits) {
			return setWithProperType(field.Type.Kind(), textSplits[i], value)
		}
		return fmt.Errorf("invalid text index")
	})
}
