package mqrr

import (
	"context"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"net/url"
	"path"
	"reflect"
	"runtime"
	"strings"
	"time"
)

// Engine is the server instance, it contains the connection manager, router and subscriptions.
// Create an instance of Engine, by using New().
type Engine struct {
	RouterGroup
	BaseTopic     string
	subscriptions map[string]paho.SubscribeOptions
	router        *paho.StandardRouter
	cm            *autopaho.ConnectionManager
	handlers      map[string]interface{}
}

// New returns a new server instance.
func New() *Engine {
	engine := &Engine{
		router:        paho.NewStandardRouter(),
		subscriptions: make(map[string]paho.SubscribeOptions),
		handlers:      make(map[string]interface{}),
	}
	engine.RouterGroup.engine = engine
	return engine
}

// Route registers a request handler with the given topic.
// A topic contains multiple levels, each level is separated by a forward slash.
// A level can be a name, or wildcards like `+` and `#`, or a named variable
// starts with `:` and `*`.
func (engine *Engine) Route(topic string, handler func(c *Context)) {
	namedTopic := path.Join(engine.BaseTopic, topic)
	absoluteTopic, params := engine.buildTopic(namedTopic)
	engine.subscriptions[absoluteTopic] = paho.SubscribeOptions{QoS: 0}
	engine.handlers[namedTopic] = handler
	engine.router.RegisterHandler(absoluteTopic, func(publish *paho.Publish) {
		go engine.handleRequest(buildContext(publish, params), handler)
	})
}

// Run connects to the given MQTT broker, then starts listening requests.
func (engine *Engine) Run(broker string) {
	brokerUrl, err := url.Parse(broker)
	if err != nil {
		panic(err)
	}
	engine.RunCfg(autopaho.ClientConfig{
		BrokerUrls: []*url.URL{brokerUrl},
		KeepAlive:  30,
	})
}

// RunUser connects to the MQTT broker using auth user and password,
// then starts listening requests.
func (engine *Engine) RunUser(broker string, user, password string) {
	brokerUrl, err := url.Parse(broker)
	if err != nil {
		panic(err)
	}
	cc := autopaho.ClientConfig{
		BrokerUrls: []*url.URL{brokerUrl},
		KeepAlive:  30,
	}
	cc.SetUsernamePassword(user, []byte(password))
	engine.RunCfg(cc)
}

// RunCfg connects to the MQTT broker using the given client config,
// then starts listening requests.
func (engine *Engine) RunCfg(cc autopaho.ClientConfig) {
	var err error
	engine.printRoute(cc.BrokerUrls)
	subs := engine.buildSubscriptions()
	if len(subs) == 0 {
		panic("no route found")
	}
	// User-defined callbacks
	onConnectionUp := cc.OnConnectionUp
	onConnectError := cc.OnConnectError
	cc.OnConnectionUp = func(manager *autopaho.ConnectionManager, connack *paho.Connack) {
		// Subscribe all the registered topics
		if _, err := manager.Subscribe(context.Background(), &paho.Subscribe{Subscriptions: subs}); err != nil {
			log.Error(err)
		}
		if onConnectionUp != nil {
			onConnectionUp(manager, connack)
		}
	}
	cc.OnConnectError = func(err error) {
		log.Error(err)
		if onConnectError != nil {
			onConnectError(err)
		}
	}
	cc.ClientConfig.Router = engine.router
	// Start making connection to the broker
	engine.cm, err = autopaho.NewConnection(context.Background(), cc)
	if err != nil {
		panic(err)
	}
	// Wait for the connection manager to exit
	<-engine.cm.Done()
}

func (engine *Engine) buildTopic(s string) (string, map[string]int) {
	levels := make([]string, 0)
	params := make(map[string]int)
	setParam := func(key string, value int) {
		if key == "" {
			panic("invalid params in topic")
		}
		params[key] = value
	}
	topicParts := strings.Split(s, "/")
	for i, part := range topicParts {
		if part == "" {
			panic("invalid topic")
		}
		level := part
		if part[0] == ':' {
			// Single Level Wildcard
			level = "+"
			setParam(part[1:], i)
		} else if part[0] == '*' {
			// Multi Level Wildcard
			level = "#"
			if i != len(topicParts)-1 {
				panic("the multi-level wildcard must be placed as the last level in the topic")
			}
			setParam(part[1:], -i)
		}
		levels = append(levels, level)
	}
	return strings.Join(levels, "/"), params
}

// Multiple routes can share the same subscription. We should merge them
// into one subscription to prevent from receiving multiple same publish.
func (engine *Engine) buildSubscriptions() map[string]paho.SubscribeOptions {
	topics := make([]string, 0)
	for k := range engine.subscriptions {
		topics = append(topics, k)
	}
	redundant := make(map[int]bool)
	for i := range topics {
		for j := i + 1; j < len(topics); j++ {
			if redundant[i] || redundant[j] {
				continue
			}
			if match(topics[i], topics[j]) {
				redundant[j] = true
			} else if match(topics[j], topics[i]) {
				redundant[i] = true
			}
		}
	}
	subs := make(map[string]paho.SubscribeOptions)
	for i := range topics {
		if !redundant[i] {
			subs[topics[i]] = engine.subscriptions[topics[i]]
		}
	}
	return subs
}

func (engine *Engine) handleRequest(c *Context, handler func(c *Context)) {
	defer func() {
		if err := recover(); err != nil {
			log.Error(err)
		}
	}()
	// Calling handler function
	start := time.Now()
	handler(c)
	elapsed := time.Since(start)
	log.Infof("%13v | %#v", elapsed, c.Request.Topic)
	// Write response to client
	if c.Request.Properties.ResponseTopic != "" && len(c.response) > 0 {
		engine.cm.Publish(context.Background(), &paho.Publish{
			QoS:        0,
			Retain:     false,
			Topic:      c.Request.Properties.ResponseTopic,
			Payload:    c.response,
			Properties: &paho.PublishProperties{CorrelationData: c.Request.Properties.CorrelationData},
		})
	}
}

func (engine *Engine) printRoute(urls []*url.URL) {
	for topic, handler := range engine.handlers {
		debugPrint("%-25s --> %s", topic, runtime.FuncForPC(reflect.ValueOf(handler).Pointer()).Name())
	}
	debugPrint("Listening requests on %v", urls)
}

// Close closes the connection and waits for goroutine to exit.
func (engine *Engine) Close(ctx context.Context) error {
	return engine.cm.Disconnect(ctx)
}

func match(r1, r2 string) bool {
	if r1 == r2 {
		return true
	}
	return matchDeep(routeSplit(r1), routeSplit(r2))
}

func matchDeep(r1 []string, r2 []string) bool {
	if len(r1) == 0 {
		return len(r2) == 0
	}

	if len(r2) == 0 {
		return r1[0] == "#"
	}

	if r1[0] == "#" {
		return true
	}

	if (r1[0] == "+") || (r1[0] == r2[0]) {
		return matchDeep(r1[1:], r2[1:])
	}
	return false
}

func routeSplit(route string) []string {
	if len(route) == 0 {
		return nil
	}
	var result []string
	if strings.HasPrefix(route, "$share") {
		result = strings.Split(route, "/")[1:]
	} else {
		result = strings.Split(route, "/")
	}
	return result
}
