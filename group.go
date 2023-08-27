package fernet

import (
	"net/http"
	"strings"
)

type (
	// Group is a collection of routes that share a common prefix and set of middleware.
	Group[RequestData any] struct {
		prefix     string
		middleware []func(Response, *Request[RequestData], Handler[RequestData])
		parent     Routable[RequestData]
	}
)

var _ Routable[int] = (*Group[int])(nil)

// NewGroup returns a new Group instance.
func NewGroup[RequestData any](parent Routable[RequestData], prefix string) *Group[RequestData] {
	return &Group[RequestData]{
		prefix:     prefix,
		parent:     parent,
		middleware: make([]func(Response, *Request[RequestData], Handler[RequestData]), 0),
	}
}

// Match registers a route with the given method and path
func (g *Group[RequestData]) Match(method string, path string, fn Handler[RequestData]) {
	routePath := strings.TrimSuffix(g.prefix, "/") + "/" + strings.TrimPrefix(path, "/")
	g.parent.Match(method, routePath, g.wrap(fn))
}

// Get registers a GET route with the given handler
func (g *Group[RequestData]) Get(path string, fn Handler[RequestData]) {
	g.Match(http.MethodGet, path, fn)
}

// Post registers a POST route with the given handler
func (g *Group[RequestData]) Post(path string, fn Handler[RequestData]) {
	g.Match(http.MethodPost, path, fn)
}

// Put registers a PUT route with the given handler
func (g *Group[RequestData]) Put(path string, fn Handler[RequestData]) {
	g.Match(http.MethodPut, path, fn)
}

// Patch registers a PATCH route with the given handler
func (g *Group[RequestData]) Patch(path string, fn Handler[RequestData]) {
	g.Match(http.MethodPatch, path, fn)
}

// Delete registers a DELETE route with the given handler
func (g *Group[RequestData]) Delete(path string, fn Handler[RequestData]) {
	g.Match(http.MethodDelete, path, fn)
}

// Use registers a middleware that will run before the handlers of this group and subgroups.
func (g *Group[RequestData]) Use(fn func(Response, *Request[RequestData], Handler[RequestData])) {
	g.middleware = append(g.middleware, fn)
}

// Group returns a new route group with the given prefix. The group can define
// its own middleware that will only be run for that group.
func (g *Group[RequestData]) Group(prefix string) *Group[RequestData] {
	return NewGroup[RequestData](g, prefix)
}

// wrap takes a Handler and ensures that this groups middleware is run before the handler is called
func (g *Group[RequestData]) wrap(fn Handler[RequestData]) Handler[RequestData] {
	return func(res Response, req *Request[RequestData]) {
		handler := fn

		for i := len(g.middleware) - 1; i >= 0; i-- {
			currentHandler := handler
			middleware := g.middleware[i]

			handler = func(res Response, req *Request[RequestData]) {
				middleware(res, req, currentHandler)
			}
		}

		handler(res, req)
	}
}
