package binder

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var jsonTestData = []byte(`{"name":"john","age":50,"sex":"male"}`)

type InnerBindingTest struct {
	Location string `json:"location"`
	Salary   int    `json:"salary" text:"3,,"`
	X        *int   `text:"1,,"`
}

type bindingTest struct {
	Name  string   `json:"name" text:"0,," topic:"name"`
	Age   int      `json:"age" text:"1,," topic:"age"`
	Text  string   `text:"0" topic:"last"`
	Slice []string `topic:"last"`
	InnerBindingTest
}

func TestJsonBinderBind(t *testing.T) {
	obj := bindingTest{}
	require.NoError(t, JSON.Bind(jsonTestData, &obj))
	assert.Equal(t, "john", obj.Name)
	assert.Equal(t, 50, obj.Age)
}

func TestJSONUseNumber(t *testing.T) {
	JSONUseNumber = true
	var m map[string]interface{}
	require.NoError(t, JSON.Bind([]byte(`{"a":1,"b":1.2}`), &m))
	assert.Equal(t, json.Number("1"), m["a"])
	assert.Equal(t, json.Number("1.2"), m["b"])
}

func TestJSONDisallowUnknownFields(t *testing.T) {
	JSONDisallowUnknownFields = true
	obj := bindingTest{}
	require.Error(t, JSON.Bind(jsonTestData, &obj))
}

var textTestData = []byte(`john,50,male,10000`)

func TestTextBinderBind(t *testing.T) {
	obj := bindingTest{}
	require.NoError(t, Text.Bind(textTestData, &obj))
	x := 50
	assert.Equal(t, bindingTest{
		Name: "john",
		Age:  50,
		Text: "john,50,male,10000",
		InnerBindingTest: InnerBindingTest{
			Salary: 10000,
			X:      &x,
		},
	}, obj)
}

var topicTestData = map[string][]string{
	"name": {"john"},
	"age":  {"50"},
	"last": {"male", "tester"},
}

func TestTopicBinderBind(t *testing.T) {
	obj := bindingTest{}
	require.NoError(t, Topic.Bind(topicTestData, &obj))
	assert.Equal(t, bindingTest{
		Name:  "john",
		Age:   50,
		Text:  "male/tester",
		Slice: topicTestData["last"],
	}, obj)
}
