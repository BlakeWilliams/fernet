package fernet

import (
	"context"
	"net/http"

	"github.com/blakewilliams/fernet/internal/radical"
)

type (
	Middleware[T RequestContext] func(context.Context, T, Handler[T])

	Handler[T RequestContext] func(context.Context, T)

	// Router represents the primary router for the application.
	Router[T RequestContext] struct {
		routes     []*route[T]
		tree       *radical.Node[*route[T]]
		metal      *MetalStack[T]
		middleware []func(context.Context, T, Handler[T])
		initT      func(RequestContext) T
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
		Use(func(context.Context, T, Handler[T]))

		// Group returns a new group based on this Routable. It will have its
		// own middleware stack in addition to the middleware stack on the
		// groups/router above it.
		Group(prefix string) *Group[T]
	}
)

var _ Routable[*RootRequestContext] = (*Router[*RootRequestContext])(nil)

// New returns a new router with the given RequestContext type. The function
// passed to this function is used to initialize the RequestContext for each
// request which is then passed to the relevant route handler.
func New[T RequestContext, Init func(RequestContext) T](init Init) *Router[T] {
	r := &Router[T]{
		tree:       radical.New[*route[T]](),
		middleware: make([]func(context.Context, T, Handler[T]), 0),
		initT:      init,
	}

	r.metal = NewMetalStack[T](http.HandlerFunc(r.handler))

	return r
}

// Match registers a route with the router.
func (r *Router[T]) Match(method string, path string, handler Handler[T]) {
	route := newRoute[T](method, path, handler)
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

// Use registers a middleware that will be run after the UseMetal middleware but
// before the handler. Each middleware is passed the next Handler or middleware
// in the stack. Not calling the next function will halt the middleware/handler chain.
func (r *Router[T]) Use(fn func(context.Context, T, Handler[T])) {
	r.middleware = append(r.middleware, fn)
}

// Group returns a new route group with the given prefix. The group can define
// its own middleware that will only be run for that group.
func (r *Router[T]) Group(prefix string) *Group[T] {
	return NewGroup[T](r, prefix)
}

// ServeHTTP implements the http.Handler interface.
func (r *Router[T]) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	r.metal.ServeHTTP(rw, req)
}

// Metal returns the MetalStack for this router which can be used to register
// http based middleware that will run before the fernet middleware and
// handlers.
func (r *Router[T]) Metal() *MetalStack[T] {
	return r.metal
}

func (r *Router[T]) handler(rw http.ResponseWriter, req *http.Request) {
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
		if !ok {
			// This should never actually get hit in real code but would
			// indicate a bug in the framework.
			panic("route did not match request")
		}
	} else {
		params = map[string]string{}
		handler = func(ctx context.Context, rctx T) {
			rctx.Response().WriteHeader(http.StatusNotFound)
		}
	}

	for i := len(r.middleware) - 1; i >= 0; i-- {
		currentHandler := handler
		middleware := r.middleware[i]
		handler = func(ctx context.Context, reqCtx T) {
			middleware(ctx, reqCtx, currentHandler)
		}
	}

	res := newResponseWriter(rw)
	reqCtx := newRequestContext(req, res, path, params)
	handler(
		req.Context(),
		r.initT(reqCtx),
	)

	res.Flush()
}

type Registerable[T RequestContext] interface {
	Register(Routable[T])
}

// Register passes this routable to the provided registerable so that it can
// register its own routes on the routable. This is useful for building
// abstractions like controllers or packages that need to manage and register
// their own routes/state.
func (r *Router[T]) Register(c Registerable[T]) {
	c.Register(r)
}
