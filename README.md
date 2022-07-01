# MQRR

MQTT v5 Request-Response (MQRR) Pattern provides similar behavior like HTTP. 
Different from the HTTP Request-Response model, MQTT request-response is asynchronous, 
which brings a problem, that is how to associate the response message with the request message.
The following diagram describes the request-response interaction process:

![img](https://assets.emqx.com/images/d624fb3a3061f043f32ae02338f635a0.png?imageMogr2/thumbnail/1520x)

See the [MQTT Request Response](https://www.emqx.com/en/blog/mqtt5-request-response) article for detailed description.

## Installation
```shell
go get github.com/koho/mqrr
```

## Quick start

### Server

Let's create a simple API server that listens for the `hello` topic.

```go
package main

import (
	"github.com/koho/mqrr"
)

func main() {
	r := mqrr.New()
	r.Route("hello", func(c *mqrr.Context) {
		c.String("Hello %s", c.GetRawString())
	})
	r.Run("mqtt://broker-cn.emqx.io:1883")
}
```

### Client

In the client side, we publish a message with our name to the `hello` topic, then wait for a response.

```shell
go get github.com/koho/mqrr/client
```

```go
package main

import (
	"context"
	"fmt"
	"github.com/eclipse/paho.golang/paho"
	"github.com/koho/mqrr/client"
)

func main() {
	resp, err := client.Request(context.Background(), "mqtt://broker-cn.emqx.io:1883", &paho.Publish{
		Topic:   "hello",
		Payload: []byte("John"),
	})
	if err != nil {
		panic(err)
	}
	fmt.Println(string(resp.Payload)) // Should output "Hello John"
}
```

## API Examples

### Parameters in topic
```go
func main() {
	r := mqrr.New()
	r.Route("user/:name", func(c *mqrr.Context) {
        name := c.Param("name")
		c.String("Hello %s", name)
	})
	r.Route("groups/*last", func(c *mqrr.Context) {
		c.String(c.Param("last"))
	})
	r.Run("mqtt://broker-cn.emqx.io:1883")
}
```

### Grouping routes
```go
func main() {
	r := mqrr.New()
	g1 := r.Group("G1")
	g1.Route(":dev/new", func(c *mqrr.Context) {
		c.String("new")
	})
	g2 := g1.Group("G2")
	g2.Route(":temp", func(c *mqrr.Context) {
		c.String("temp")
	})
	r.Run("mqtt://broker-cn.emqx.io:1883")
}
```

### Data binding
```go
type User struct {
	Name string `topic:"name"`
	Age  int    `json:"age" validate:"gte=18"`
}

func main() {
	r := mqrr.New()
	r.Route("user/:name", func(c *mqrr.Context) {
		user := User{}
		if err := c.BindTopic(&user); err != nil {
			c.String(err.Error())
			return
		}
		if err := c.ShouldBindJSON(&user); err != nil {
			c.String(err.Error())
			return
		}
		c.JSON(user)
	})
	r.Run("mqtt://broker-cn.emqx.io:1883")
}
```

### Client requests in same connection
```go
func main() {
	c, err := client.New("mqtt://broker-cn.emqx.io:1883")
	if err != nil {
		panic(err)
	}
	defer c.Close(context.Background())
	wg := sync.WaitGroup{}
	for _, name := range []string{"John", "Mary", "Ben"} {
		wg.Add(1)
		go func(n string) {
			defer wg.Done()
			resp, err := c.Request(context.Background(), &paho.Publish{
				Topic:   "hello",
				Payload: []byte(n),
			})
			if err != nil {
				panic(err)
			}
			fmt.Println(string(resp.Payload))
		}(name)
	}
	wg.Wait()
}
```
