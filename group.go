package mqrr

import "path"

// RouterGroup is associated with a topic prefix.
// In the Route call, it joins all the topic levels to form a full topic.
type RouterGroup struct {
	engine *Engine
	base   string
}

// Group creates a new router group with the given topic prefix.
func (g *RouterGroup) Group(base string) *RouterGroup {
	return &RouterGroup{
		engine: g.engine,
		base:   path.Join(g.base, base),
	}
}

// Route registers a request handler with the given topic.
// See Engine.Route for detail.
func (g *RouterGroup) Route(topic string, handler func(c *Context)) {
	g.engine.Route(path.Join(g.base, topic), handler)
}
