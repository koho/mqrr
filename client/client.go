package client

import (
	"context"
	"fmt"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"net/url"
	"sync"
)

// Client provides a long connection to the MQTT broker,
// so that you can make continuous requests at the same time.
type Client struct {
	sync.RWMutex
	cm      *autopaho.ConnectionManager
	connUp  chan struct{}
	handler *Handler
	router  paho.Router
}

// New creates a default Client with the given broker url.
func New(broker string) (*Client, error) {
	brokerUrl, err := url.Parse(broker)
	if err != nil {
		return nil, err
	}
	cc := autopaho.ClientConfig{
		BrokerUrls: []*url.URL{brokerUrl},
		KeepAlive:  30,
	}
	return NewWithCfg(cc)
}

// NewWithCfg creates a new Client with the given client config.
func NewWithCfg(cc autopaho.ClientConfig) (*Client, error) {
	client := &Client{
		router: paho.NewSingleHandlerRouter(nil),
		connUp: make(chan struct{}),
	}
	cc.OnConnectionUp = client.onConnectionUp
	cc.OnConnectError = client.onConnectionError
	cc.ClientConfig.Router = client.router
	var err error
	client.cm, err = autopaho.NewConnection(context.Background(), cc)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func (client *Client) onConnectionUp(manager *autopaho.ConnectionManager, connack *paho.Connack) {
	client.Lock()
	defer client.Unlock()
	client.handler, _ = NewHandler(manager, client.router, connack.Properties.AssignedClientID)
	close(client.connUp)
}

func (client *Client) onConnectionError(err error) {
	client.Lock()
	defer client.Unlock()
	client.handler = nil
	client.connUp = make(chan struct{})
}

// Request sends a request to the MQTT broker and waits for a response.
func (client *Client) Request(ctx context.Context, pb *paho.Publish) (*paho.Publish, error) {
	client.RLock()
	ch := client.connUp
	client.RUnlock()
	// Wait for the connection up
	select {
	case <-ch:
	case <-ctx.Done():
		return nil, ctx.Err()
	}
	// Make request
	client.RLock()
	defer client.RUnlock()
	if client.handler == nil {
		return nil, fmt.Errorf("mqtt not connected")
	}
	return client.handler.Request(ctx, pb)
}

// Close disconnects the Client and waits for the connection manager to exit.
func (client *Client) Close(ctx context.Context) error {
	return client.cm.Disconnect(ctx)
}
