package mqrr

import (
	"github.com/eclipse/paho.golang/paho"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type contextBinding struct {
	Name string `json:"name" validate:"required" text:"0,-" topic:"name"`
	Age  int    `json:"age" validate:"gte=40" text:"1,-" topic:"age"`
}

var topicContext = buildContext(&paho.Publish{Topic: "test/john/50/sex/a/b/c"}, map[string]int{
	"name": 1,
	"age":  2,
	"last": -4,
})

func TestContextShouldBindJSON(t *testing.T) {
	ctx := buildContext(&paho.Publish{Payload: []byte(`{"name":"john","age":50}`)}, nil)
	obj := contextBinding{}
	require.NoError(t, ctx.ShouldBindJSON(&obj))
	assert.Equal(t, contextBinding{"john", 50}, obj)

	ctx = buildContext(&paho.Publish{Payload: []byte(`{"name":"","age":30}`)}, nil)
	require.Error(t, ctx.ShouldBindJSON(&obj))
}

func TestContextShouldBindText(t *testing.T) {
	ctx := buildContext(&paho.Publish{Payload: []byte(`john-50-96-10`)}, nil)
	obj := contextBinding{}
	require.NoError(t, ctx.ShouldBindText(&obj))
	assert.Equal(t, contextBinding{"john", 50}, obj)
}

func TestContextShouldBindTopic(t *testing.T) {
	obj := contextBinding{}
	require.NoError(t, topicContext.ShouldBindTopic(&obj))
	assert.Equal(t, contextBinding{"john", 50}, obj)
}

func TestContextJSON(t *testing.T) {
	ctx := buildContext(&paho.Publish{}, nil)
	obj := struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{"john", 50}
	ctx.JSON(obj)
	assert.Equal(t, []byte(`{"name":"john","age":50}`), ctx.response)
}

func TestContextString(t *testing.T) {
	ctx := buildContext(&paho.Publish{}, nil)
	ctx.String("hello")
	assert.Equal(t, []byte("hello"), ctx.response)
}

func TestContextData(t *testing.T) {
	ctx := buildContext(&paho.Publish{}, nil)
	ctx.Data([]byte("hello"))
	assert.Equal(t, []byte("hello"), ctx.response)
}

func TestContextParam(t *testing.T) {
	assert.Equal(t, "john", topicContext.Param("name"))
	assert.Equal(t, "50", topicContext.Param("age"))
	assert.Equal(t, "a/b/c", topicContext.Param("last"))
}
