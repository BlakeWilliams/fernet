package fernet

import (
	"context"
	"net/http"
	"strings"

	"github.com/blakewilliams/fernet/internal/radical"
)

type (
	// Middleware is a function that wraps a handler and other middlewares. They
	// accept a context, the RequestContext, and the next handler to be called.
	// If the `next` handler is not called, the request halts.
	Middleware[T RequestContext] func(context.Context, T, Handler[T])

	// Handler is a function that handles a request.
	Handler[T RequestContext] func(context.Context, T)

	// Router represents the primary router for the application.
	Router[T RequestContext] struct {
		routes           []*route[T]
		tree             *radical.Node[*route[T]]
		middleware       []func(context.Context, T, Handler[T])
		metal            []func(w http.ResponseWriter, r *http.Request, next http.Handler)
		initT            func(RequestContext) T
		anyRoutesDefined bool
	}

	// Registerable is an interface that can be implemented by types that want
	// to register routes with a router. This allows the router to be extended
	// by internal or external packages like Group, and Controller.
	Registerable[T RequestContext] interface {
		// RawMatch registers a route with the given method and path
		RawMatch(method string, path string, fn Handler[T])
	}

	// Routable is an interface that can be implemented by types that want to
	// register routes with a router.
	Routable[T RequestContext] interface {
		// Match registers a route with the given method and path
		Match(method string, path string, fn Handler[T])
		// Get registers a GET route with the given path
		Get(method string, fn Handler[T])
		// Post registers a POST route with the given path
		Post(method string, fn Handler[T])
		// Put registers a PUT route with the given path
		Put(method string, fn Handler[T])
		// Patch registers a PATCH route with the given path
		Patch(method string, fn Handler[T])
		// Delete registers a DELETE route with the given path
		Delete(method string, fn Handler[T])

		// Use registers a middleware function that is run before each request
		// for this group and all groups below it.
		Use(...func(context.Context, T, Handler[T]))

		// Group returns a new group based on this Routable. It will have its
		// own middleware stack in addition to the middleware stack on the
		// groups/router above it.
		Group() *Group[T]

		// Namespace is like Group but accepts a prefix that will be included in
		// the path of all routes registered with the group.
		Namespace(prefix string) *Group[T]
	}
)

var _ Routable[*RootRequestContext] = (*Router[*RootRequestContext])(nil)
var _ Registerable[*RootRequestContext] = (*Router[*RootRequestContext])(nil)

// New returns a new router with the given RequestContext type. The function
// passed to this function is used to initialize the RequestContext for each
// request which is then passed to the relevant route handler.
func New[T RequestContext](init func(RequestContext) T) *Router[T] {
	r := &Router[T]{
		tree:       radical.New[*route[T]](),
		middleware: make([]func(context.Context, T, Handler[T]), 0),
		initT:      init,
	}

	return r
}

// RawMatch implements the Registerable interface and registers a route with the
// router.
func (r *Router[T]) RawMatch(method string, path string, handler Handler[T]) {
	r.Match(method, path, handler)
}

// Match registers a route with the router.
func (r *Router[T]) Match(method string, path string, handler Handler[T]) {
	r.anyRoutesDefined = true

	route := newRoute[T](method, path, r.wrap(handler))
	r.routes = append(r.routes, route)

	pathParts := make([]string, 0, len(route.parts)+1)
	pathParts = append(pathParts, method)
	pathParts = append(pathParts, route.parts...)

	r.tree.Add(pathParts, route)
}

// Get registers a GET route with the router.
func (r *Router[T]) Get(path string, handler Handler[T]) {
	r.Match(http.MethodGet, path, handler)
}

// Get registers a GET route with the router.
func (r *Router[T]) Post(path string, handler Handler[T]) {
	r.Match(http.MethodPost, path, handler)
}

// Put registers a PUT route with the router.
func (r *Router[T]) Put(path string, handler Handler[T]) {
	r.Match(http.MethodPut, path, handler)
}

// Patch registers a PATCH route with the router.
func (r *Router[T]) Patch(path string, handler Handler[T]) {
	r.Match(http.MethodPatch, path, handler)
}

// Delete registers a DELETE route with the router.
func (r *Router[T]) Delete(path string, handler Handler[T]) {
	r.Match(http.MethodDelete, path, handler)
}

// Use registers middleware that will be run before each handler, including
// the handlers of groups and controllers.
func (r *Router[T]) Use(fns ...func(context.Context, T, Handler[T])) {
	if r.anyRoutesDefined {
		panic("Use can only be called before routes are defined")
	}

	r.middleware = append(r.middleware, fns...)
}

// UseMetal registers "metal" middleware (net/http based) that will be run
// before the fernet middleware stack and route handler. This is useful for
// when the underlying http.ResponseWriter or *http.Request need to be
// modified before fernet uses them.
func (r *Router[T]) UseMetal(fns ...func(w http.ResponseWriter, r *http.Request, next http.Handler)) {
	if r.anyRoutesDefined {
		panic("UseMetal can only be called before routes are defined")
	}

	r.metal = append(r.metal, fns...)
}

// Group returns a new route group that can define its own middleware
// that will only be run for that group.
func (r *Router[T]) Group() *Group[T] {
	return NewGroup[T](r, "")
}

// Namespace returns a new route group prefix. The group can define its own
// middleware that will only be run for that group.
func (r *Router[T]) Namespace(prefix string) *Group[T] {
	return NewGroup[T](r, prefix)
}

// ServeHTTP implements the http.Handler interface.
func (r *Router[T]) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	httpHandler := func(rw http.ResponseWriter, req *http.Request) {
		// Run fernet middleware and call route handler
		method := req.Method
		normalizedPath := normalizeRoutePath(req.URL.Path)
		lookup := []string{method}
		lookup = append(lookup, normalizedPath...)

		var handler func(context.Context, T)
		var params map[string]string
		var path string

		ok, value := r.tree.Value(lookup)
		if ok {
			handler = value.handler
			path = value.Path

			var ok bool
			ok, params = value.match(req)
			if !ok && !value.isWildcard() {
				// This should never actually get hit in real code but would
				// indicate a bug in the framework.
				panic("route did not match request. this is a bug in fernet. please open an issue reporting this error and how to reproduce it.")
			}
		} else {
			params = map[string]string{}
			handler = r.wrap(func(ctx context.Context, rctx T) {
				rctx.Response().WriteHeader(http.StatusNotFound)
			})
		}

		reqCtx := NewRequestContext(req, rw, path, params)
		handler(
			req.Context(),
			r.initT(reqCtx),
		)

		reqCtx.Response().Flush()
	}

	for i := len(r.metal) - 1; i >= 0; i-- {
		currentHandler := httpHandler
		m := r.metal[i]

		httpHandler = http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
			m(rw, req, http.HandlerFunc(currentHandler))
		})
	}

	httpHandler(rw, req)
}

func (r *Router[T]) wrap(fn Handler[T]) func(context.Context, T) {
	handler := fn

	for i := len(r.middleware) - 1; i >= 0; i-- {
		currentHandler := handler
		middleware := r.middleware[i]
		handler = func(ctx context.Context, reqCtx T) {
			middleware(ctx, reqCtx, currentHandler)
		}
	}

	return handler
}

func joinURL(prefix string, path string) string {
	if prefix == "" {
		return path
	}

	if path == "" {
		return prefix
	}

	if path == "/" {
		return prefix
	}

	return strings.TrimSuffix(prefix, "/") + "/" + strings.TrimPrefix(path, "/")
}
