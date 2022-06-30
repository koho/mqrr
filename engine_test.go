package mqrr

import (
	"context"
	"github.com/eclipse/paho.golang/paho"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var broker = "mqtt://broker-cn.emqx.io:1883"

func TestEngineRoute(t *testing.T) {
	r := New()
	r.BaseTopic = "MQRR"
	r.Route(":name/:age/*last", func(c *Context) { println("LAST") })
	r.Route("tests/test", func(c *Context) { println("TEST") })
	g1 := r.Group("G1")
	g1.Route(":name/new", func(c *Context) {})
	g2 := g1.Group("G2")
	g2.Route(":temp", func(c *Context) {})
	assert.Contains(t, r.subscriptions, "MQRR/+/+/#")
	assert.Contains(t, r.subscriptions, "MQRR/tests/test")
	assert.Contains(t, r.subscriptions, "MQRR/G1/+/new")
	assert.Contains(t, r.subscriptions, "MQRR/G1/G2/+")
	assert.Equal(t, map[string]paho.SubscribeOptions{"MQRR/+/+/#": {}}, r.buildSubscriptions())
}

func TestEngineRun(t *testing.T) {
	r := New()
	r.Route(t.Name(), func(c *Context) {})
	go r.Run(broker)
	defer r.Close(context.Background())
	time.Sleep(2 * time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, r.engine.cm.AwaitConnection(ctx))
}
