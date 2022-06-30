package binder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// JSONUseNumber causes the Decoder to unmarshal a number into an interface{} as a
// Number instead of as a float64.
var JSONUseNumber = false

// JSONDisallowUnknownFields causes the Decoder to return an error when the destination
// is a struct and the input contains object keys which do not match any
// non-ignored, exported fields in the destination.
var JSONDisallowUnknownFields = false

type jsonBinder struct{}

func (jsonBinder) Name() string {
	return "json"
}

func (jsonBinder) Bind(raw []byte, obj interface{}) error {
	if raw == nil {
		return fmt.Errorf("invalid payload")
	}
	return decodeJSON(bytes.NewReader(raw), obj)
}

func decodeJSON(r io.Reader, obj interface{}) error {
	decoder := json.NewDecoder(r)
	if JSONUseNumber {
		decoder.UseNumber()
	}
	if JSONDisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}
	return decoder.Decode(obj)
}
