package client

import (
	"context"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func newMQTTClient() (*autopaho.ConnectionManager, paho.Router, error) {
	cc := defaultClientConfig(broker)
	cc.Router = paho.NewSingleHandlerRouter(nil)
	cm, err := autopaho.NewConnection(context.Background(), cc)
	return cm, cc.Router, err
}

func TestHandlerRequest(t *testing.T) {
	cm, router, err := newMQTTClient()
	require.NoError(t, err)
	defer cm.Disconnect(context.Background())
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, cm.AwaitConnection(ctx))
	h, err := NewHandler(cm, router, uuid.NewString())
	require.NoError(t, err)

	engine := runEchoServer(t.Name())
	defer engine.Close(context.Background())
	time.Sleep(2 * time.Second)

	resp, err := h.Request(ctx, &paho.Publish{
		Topic:   t.Name(),
		Payload: []byte(t.Name()),
	})
	require.NoError(t, err)
	assert.Equal(t, []byte(t.Name()), resp.Payload)
	h.Close(context.Background())
}
