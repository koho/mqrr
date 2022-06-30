package client

import (
	"context"
	"github.com/eclipse/paho.golang/paho"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestRequest(t *testing.T) {
	engine := runEchoServer(t.Name())
	defer engine.Close(context.Background())
	time.Sleep(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := Request(ctx, broker, &paho.Publish{
		QoS:     0,
		Topic:   t.Name(),
		Payload: []byte(t.Name()),
	})
	require.NoError(t, err)
	assert.Equal(t, []byte(t.Name()), resp.Payload)
}

func TestRequestWithCfg(t *testing.T) {
	engine := runEchoServer(t.Name())
	defer engine.Close(context.Background())
	time.Sleep(2 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	resp, err := RequestWithCfg(ctx, defaultClientConfig(broker), &paho.Publish{
		QoS:     0,
		Topic:   t.Name(),
		Payload: []byte(t.Name()),
	})
	require.NoError(t, err)
	assert.Equal(t, []byte(t.Name()), resp.Payload)
}
