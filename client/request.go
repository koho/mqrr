package client

import (
	"context"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"net/url"
	"sync"
)

type responsePub struct {
	pub *paho.Publish
	err error
}

// Request sends a request to the given MQTT broker and waits for a response.
func Request(ctx context.Context, broker string, pb *paho.Publish) (*paho.Publish, error) {
	brokerUrl, err := url.Parse(broker)
	if err != nil {
		return nil, err
	}
	cc := autopaho.ClientConfig{
		BrokerUrls: []*url.URL{brokerUrl},
		KeepAlive:  30,
	}
	return RequestWithCfg(ctx, cc, pb)
}

// RequestWithCfg connects to the MQTT broker using given config.
// After a connection is made, it sends a request to the broker and
// waits for a response.
func RequestWithCfg(ctx context.Context, cc autopaho.ClientConfig, pb *paho.Publish) (*paho.Publish, error) {
	var req sync.Once
	resp := make(chan responsePub, 1)
	router := paho.NewSingleHandlerRouter(nil)
	cc.OnConnectionUp = func(manager *autopaho.ConnectionManager, connack *paho.Connack) {
		req.Do(func() {
			if h, err := NewHandler(manager, router, connack.Properties.AssignedClientID); err == nil {
				pub, err := h.Request(ctx, pb)
				resp <- responsePub{pub, err}
			} else {
				resp <- responsePub{}
			}
		})
	}
	cc.ClientConfig.Router = router
	cm, err := autopaho.NewConnection(ctx, cc)
	if err != nil {
		return nil, err
	}
	select {
	// Wait for a response
	case r := <-resp:
		if err = cm.Disconnect(ctx); err != nil {
			return nil, err
		}
		return r.pub, r.err
	case <-cm.Done():
		return nil, ctx.Err()
	}
}
