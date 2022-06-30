package mqrr

import (
	"encoding/json"
	"fmt"
	"github.com/eclipse/paho.golang/paho"
	"github.com/koho/mqrr/binder"
	"strings"
)

// Context is a data container. It allows us to pass variables across
// different procedures, bind request data, validate struct and render
// response.
type Context struct {
	Request  *paho.Publish
	Params   map[string][]string
	response []byte
}

func buildContext(request *paho.Publish, params map[string]int) *Context {
	ctx := &Context{Request: request, Params: make(map[string][]string)}
	topicSplit := strings.Split(request.Topic, "/")
	// Build topic parameters
	for k, v := range params {
		if v < 0 {
			ctx.Params[k] = topicSplit[-v:]
		} else {
			ctx.Params[k] = []string{topicSplit[v]}
		}
	}
	return ctx
}

// Param returns the value of the topic param.
func (c *Context) Param(key string) string {
	if v, ok := c.Params[key]; ok {
		return strings.Join(v, "/")
	} else {
		return ""
	}
}

// GetRawString return raw payload data as string.
func (c *Context) GetRawString() string {
	return string(c.Request.Payload)
}

// GetRawData return raw payload data.
func (c *Context) GetRawData() []byte {
	return c.Request.Payload
}

// JSON serializes the given struct as JSON into the response data.
func (c *Context) JSON(t interface{}) {
	var err error
	if c.response, err = json.Marshal(t); err != nil {
		panic(err)
	}
}

// Data writes raw data into the response data.
func (c *Context) Data(data []byte) {
	c.response = data
}

// String writes the given string into the response data.
func (c *Context) String(format string, values ...interface{}) {
	c.response = []byte(fmt.Sprintf(format, values...))
}

// BindTopic binds the passed struct pointer using the topic parameters.
// e.g. `topic:"var1"`.
func (c *Context) BindTopic(obj interface{}) error {
	return binder.Topic.Bind(c.Params, obj)
}

// ShouldBindTopic is a combiner of BindTopic and binder.Validate.
func (c *Context) ShouldBindTopic(obj interface{}) error {
	if err := c.BindTopic(obj); err != nil {
		return err
	}
	return binder.Validate(obj)
}

// BindJSON deserializes the request payload and binds the passed struct pointer.
// e.g. `json:"var1"`.
func (c *Context) BindJSON(obj interface{}) error {
	return binder.JSON.Bind(c.Request.Payload, obj)
}

// ShouldBindJSON is a combiner of BindJSON and binder.Validate.
func (c *Context) ShouldBindJSON(obj interface{}) error {
	if err := c.BindJSON(obj); err != nil {
		return err
	}
	return binder.Validate(obj)
}

// BindText slices the request string into all substrings separated by given separator,
// then binds the passed struct pointer according to the given slice index.
// e.g. `text:"2,-"`, `text:"0"`.
func (c *Context) BindText(obj interface{}) error {
	return binder.Text.Bind(c.Request.Payload, obj)
}

// ShouldBindText is a combiner of BindText and binder.Validate.
func (c *Context) ShouldBindText(obj interface{}) error {
	if err := c.BindText(obj); err != nil {
		return err
	}
	return binder.Validate(obj)
}
