package fernet

import (
	"context"
	"net/http"
	"strings"
)

type (
	// Group is a collection of routes that share a common prefix and set of middleware.
	Group[ReqCtx ReqRes] struct {
		prefix     string
		middleware []func(context.Context, ReqCtx, Handler[ReqCtx])
		parent     Routable[ReqCtx]
	}
)

var _ Routable[*BasicReqContext] = (*Group[*BasicReqContext])(nil)

// NewGroup returns a new Group instance.
func NewGroup[ReqCtx ReqRes](parent Routable[ReqCtx], prefix string) *Group[ReqCtx] {
	return &Group[ReqCtx]{
		prefix:     prefix,
		parent:     parent,
		middleware: make([]func(context.Context, ReqCtx, Handler[ReqCtx]), 0),
	}
}

// Match registers a route with the given method and path
func (g *Group[ReqCtx]) Match(method string, path string, fn Handler[ReqCtx]) {
	routePath := strings.TrimSuffix(g.prefix, "/") + "/" + strings.TrimPrefix(path, "/")
	g.parent.Match(method, routePath, g.wrap(fn))
}

// Get registers a GET route with the given handler
func (g *Group[ReqCtx]) Get(path string, fn Handler[ReqCtx]) {
	g.Match(http.MethodGet, path, fn)
}

// Post registers a POST route with the given handler
func (g *Group[ReqCtx]) Post(path string, fn Handler[ReqCtx]) {
	g.Match(http.MethodPost, path, fn)
}

// Put registers a PUT route with the given handler
func (g *Group[ReqCtx]) Put(path string, fn Handler[ReqCtx]) {
	g.Match(http.MethodPut, path, fn)
}

// Patch registers a PATCH route with the given handler
func (g *Group[ReqCtx]) Patch(path string, fn Handler[ReqCtx]) {
	g.Match(http.MethodPatch, path, fn)
}

// Delete registers a DELETE route with the given handler
func (g *Group[ReqCtx]) Delete(path string, fn Handler[ReqCtx]) {
	g.Match(http.MethodDelete, path, fn)
}

// Use registers a middleware that will run before the handlers of this group and subgroups.
func (g *Group[ReqCtx]) Use(fn func(context.Context, ReqCtx, Handler[ReqCtx])) {
	g.middleware = append(g.middleware, fn)
}

// Group returns a new route group with the given prefix. The group can define
// its own middleware that will only be run for that group.
func (g *Group[ReqCtx]) Group(prefix string) *Group[ReqCtx] {
	return NewGroup[ReqCtx](g, prefix)
}

// wrap takes a Handler and ensures that this groups middleware is run before the handler is called
func (g *Group[ReqCtx]) wrap(fn Handler[ReqCtx]) Handler[ReqCtx] {
	return func(ctx context.Context, r ReqCtx) {
		handler := fn

		for i := len(g.middleware) - 1; i >= 0; i-- {
			currentHandler := handler
			middleware := g.middleware[i]

			handler = func(ctx context.Context, r ReqCtx) {
				middleware(ctx, r, currentHandler)
			}
		}

		handler(ctx, r)
	}
}
