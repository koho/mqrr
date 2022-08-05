package client

import (
	"context"
	"fmt"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/google/uuid"
	"sync"
)

// Handler is the struct providing a request/response functionality
// for the paho MQTT v5 client.
type Handler struct {
	sync.Mutex
	c          *autopaho.ConnectionManager
	router     paho.Router
	respTopic  string
	correlData map[string]chan *paho.Publish
}

// NewHandler registers a response topic and listens for responses for all requests.
func NewHandler(c *autopaho.ConnectionManager, router paho.Router) *Handler {
	h := &Handler{
		c:          c,
		router:     router,
		respTopic:  fmt.Sprintf("%s/responses", uuid.NewString()),
		correlData: make(map[string]chan *paho.Publish),
	}
	router.RegisterHandler(h.respTopic, h.responseHandler)
	return h
}

// Subscribe makes a subscription to the response topic.
func (h *Handler) Subscribe(ctx context.Context) error {
	_, err := h.c.Subscribe(ctx, &paho.Subscribe{
		Subscriptions: map[string]paho.SubscribeOptions{
			h.respTopic: {QoS: 1},
		},
	})
	return err
}

func (h *Handler) addCorrelID(cID string, r chan *paho.Publish) {
	h.Lock()
	defer h.Unlock()
	h.correlData[cID] = r
}

func (h *Handler) getCorrelIDChan(cID string) chan *paho.Publish {
	h.Lock()
	defer h.Unlock()
	rChan, ok := h.correlData[cID]
	if ok {
		delete(h.correlData, cID)
	}
	return rChan
}

// Request sends a request to the MQTT broker and waits for a response.
func (h *Handler) Request(ctx context.Context, pb *paho.Publish) (*paho.Publish, error) {
	cID := uuid.NewString()
	rChan := make(chan *paho.Publish, 1)

	h.addCorrelID(cID, rChan)

	if pb.Properties == nil {
		pb.Properties = &paho.PublishProperties{}
	}

	pb.Properties.CorrelationData = []byte(cID)
	pb.Properties.ResponseTopic = h.respTopic
	pb.Retain = false

	if _, err := h.c.Publish(ctx, pb); err != nil {
		return nil, err
	}

	select {
	case resp := <-rChan:
		return resp, nil
	case <-ctx.Done():
		h.getCorrelIDChan(cID)
		return nil, ctx.Err()
	}
}

func (h *Handler) responseHandler(pb *paho.Publish) {
	if pb.Properties == nil || pb.Properties.CorrelationData == nil {
		return
	}
	rChan := h.getCorrelIDChan(string(pb.Properties.CorrelationData))
	if rChan == nil {
		return
	}
	rChan <- pb
}

// Close unregisters handlers of the response topic.
func (h *Handler) Close(ctx context.Context) error {
	if _, err := h.c.Unsubscribe(ctx, &paho.Unsubscribe{
		Topics: []string{h.respTopic},
	}); err != nil {
		return err
	}
	h.router.UnregisterHandler(h.respTopic)
	return nil
}
