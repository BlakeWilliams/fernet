package fernet

import (
	"context"
	"net/http"
	"strings"
)

type (
	// Group is a collection of routes that share a common prefix and set of middleware.
	Group[T RequestContext] struct {
		prefix     string
		middleware []func(context.Context, T, Handler[T])
		parent     Routable[T]
	}
)

var _ Routable[*RootRequestContext] = (*Group[*RootRequestContext])(nil)

// NewGroup returns a new Group instance.
func NewGroup[T RequestContext](parent Routable[T], prefix string) *Group[T] {
	return &Group[T]{
		prefix:     prefix,
		parent:     parent,
		middleware: make([]func(context.Context, T, Handler[T]), 0),
	}
}

// Match registers a route with the given method and path
func (g *Group[T]) Match(method string, path string, fn Handler[T]) {
	routePath := strings.TrimSuffix(g.prefix, "/") + "/" + strings.TrimPrefix(path, "/")
	g.parent.Match(method, routePath, g.wrap(fn))
}

// Get registers a GET route with the given handler
func (g *Group[T]) Get(path string, fn Handler[T]) {
	g.Match(http.MethodGet, path, fn)
}

// Post registers a POST route with the given handler
func (g *Group[T]) Post(path string, fn Handler[T]) {
	g.Match(http.MethodPost, path, fn)
}

// Put registers a PUT route with the given handler
func (g *Group[T]) Put(path string, fn Handler[T]) {
	g.Match(http.MethodPut, path, fn)
}

// Patch registers a PATCH route with the given handler
func (g *Group[T]) Patch(path string, fn Handler[T]) {
	g.Match(http.MethodPatch, path, fn)
}

// Delete registers a DELETE route with the given handler
func (g *Group[T]) Delete(path string, fn Handler[T]) {
	g.Match(http.MethodDelete, path, fn)
}

// Use registers a middleware that will run before the handlers of this group and subgroups.
func (g *Group[T]) Use(fn func(context.Context, T, Handler[T])) {
	g.middleware = append(g.middleware, fn)
}

// Group returns a new route group with the given prefix. The group can define
// its own middleware that will only be run for that group.
func (g *Group[T]) Group(prefix string) *Group[T] {
	return NewGroup[T](g, prefix)
}

// wrap takes a Handler and ensures that this groups middleware is run before the handler is called
func (g *Group[T]) wrap(fn Handler[T]) Handler[T] {
	return func(ctx context.Context, r T) {
		handler := fn

		for i := len(g.middleware) - 1; i >= 0; i-- {
			currentHandler := handler
			middleware := g.middleware[i]

			handler = func(ctx context.Context, r T) {
				middleware(ctx, r, currentHandler)
			}
		}

		handler(ctx, r)
	}
}
