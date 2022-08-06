package client

import (
	"context"
	"fmt"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/koho/mqrr"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"sync"
	"testing"
	"time"
)

var broker = "mqtt://broker-cn.emqx.io:1883"

func runEchoServer(topic string) *mqrr.Engine {
	r := mqrr.New()
	r.Route(topic, func(c *mqrr.Context) {
		c.String(c.GetRawString())
	})
	go r.Run(broker)
	return r
}

func defaultClientConfig(urls ...string) autopaho.ClientConfig {
	urlList := make([]*url.URL, 0)
	for _, u := range urls {
		if o, err := url.Parse(u); err == nil {
			urlList = append(urlList, o)
		}
	}
	return autopaho.ClientConfig{
		BrokerUrls: urlList,
		KeepAlive:  10,
	}
}

func TestNew(t *testing.T) {
	c, err := New(broker)
	require.NoError(t, err)
	err = c.cm.AwaitConnection(context.Background())
	require.NoError(t, err)
	defer c.Close(context.Background())
}

func TestNewWithUser(t *testing.T) {
	c, err := NewWithUser(broker, "test", "test")
	require.NoError(t, err)
	err = c.cm.AwaitConnection(context.Background())
	require.NoError(t, err)
	defer c.Close(context.Background())
}

func TestNewWithCfg(t *testing.T) {
	cc := defaultClientConfig(broker, "ws://broker-cn.emqx.io:8083")
	assert.Equal(t, 2, len(cc.BrokerUrls))
	c, err := NewWithCfg(cc)
	require.NoError(t, err)
	defer c.Close(context.Background())
}

func TestClientRequest(t *testing.T) {
	topic := "MQRR/Test"
	engine := runEchoServer(topic)
	defer engine.Close(context.Background())
	time.Sleep(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	client, err := New(broker)
	require.NoError(t, err)
	defer client.Close(context.Background())

	wg := sync.WaitGroup{}
	for i := range make([]struct{}, 5) {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			payload := []byte(fmt.Sprintf("%d", id))
			resp, err := client.Request(ctx, &paho.Publish{
				Topic:   topic,
				Payload: payload,
			})
			require.NoError(t, err)
			assert.Equal(t, resp.Payload, payload)
		}(i)
	}
	wg.Wait()
}
