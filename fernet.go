package fernet

import (
	"context"
	"net/http"

	"github.com/blakewilliams/fernet/internal/radical"
)

type (
	Handler[ReqCtx RequestContext] func(context.Context, ReqCtx)

	// Router represents the primary router for the application.
	Router[ReqCtx RequestContext] struct {
		routes     []*route[ReqCtx]
		tree       *radical.Node[*route[ReqCtx]]
		metal      []func(http.ResponseWriter, *http.Request, http.Handler)
		middleware []func(context.Context, ReqCtx, Handler[ReqCtx])
		initReqCtx func(RequestContext) ReqCtx
	}

	// Routable is an interface that can be implemented by types that want to
	// register routes with a router.
	Routable[ReqCtx RequestContext] interface {
		// Match registers a route with the given method and path
		Match(method string, path string, fn Handler[ReqCtx])
		// Get registers a GET route with the given path
		Get(method string, fn Handler[ReqCtx])
		// Post registers a POST route with the given path
		Post(method string, fn Handler[ReqCtx])
		// Put registers a PUT route with the given path
		Put(method string, fn Handler[ReqCtx])
		// Patch registers a PATCH route with the given path
		Patch(method string, fn Handler[ReqCtx])
		// Delete registers a DELETE route with the given path
		Delete(method string, fn Handler[ReqCtx])

		// Use registers a middleware function that is run before each request
		// for this group and all groups below it.
		Use(func(context.Context, ReqCtx, Handler[ReqCtx]))

		// Group returns a new group based on this Routable. It will have its
		// own middleware stack in addition to the middleware stack on the
		// groups/router above it.
		Group(prefix string) *Group[ReqCtx]
	}
)

var _ Routable[*RootRequestContext] = (*Router[*RootRequestContext])(nil)

func WithBasicRequestContext(rctx RequestContext) *RootRequestContext {
	return rctx.(*RootRequestContext)
}

// New returns a new Router. The provided function is used to create a new
// request context for each request. The context can be used to store data
// that should be available to all handlers in the request like the current
// user, database connections, etc.
func New[ReqCtx RequestContext, Init func(RequestContext) ReqCtx](init Init) *Router[ReqCtx] {
	return &Router[ReqCtx]{
		tree:       radical.New[*route[ReqCtx]](),
		routes:     make([]*route[ReqCtx], 0),
		metal:      make([]func(http.ResponseWriter, *http.Request, http.Handler), 0),
		middleware: make([]func(context.Context, ReqCtx, Handler[ReqCtx]), 0),
		initReqCtx: init,
	}
}

// Match registers a route with the router.
func (r *Router[ReqCtx]) Match(method string, path string, handler Handler[ReqCtx]) {
	route := newRoute[ReqCtx](method, path, handler)
	r.routes = append(r.routes, route)

	pathParts := make([]string, 0, len(route.parts)+1)
	pathParts = append(pathParts, method)
	pathParts = append(pathParts, route.parts...)

	r.tree.Add(pathParts, route)
}

// Get registers a GET route with the router.
func (r *Router[ReqCtx]) Get(path string, handler Handler[ReqCtx]) {
	r.Match(http.MethodGet, path, handler)
}

// Get registers a GET route with the router.
func (r *Router[ReqCtx]) Post(path string, handler Handler[ReqCtx]) {
	r.Match(http.MethodPost, path, handler)
}

// Put registers a PUT route with the router.
func (r *Router[ReqCtx]) Put(path string, handler Handler[ReqCtx]) {
	r.Match(http.MethodPut, path, handler)
}

// Patch registers a PATCH route with the router.
func (r *Router[ReqCtx]) Patch(path string, handler Handler[ReqCtx]) {
	r.Match(http.MethodPatch, path, handler)
}

// Delete registers a DELETE route with the router.
func (r *Router[ReqCtx]) Delete(path string, handler Handler[ReqCtx]) {
	r.Match(http.MethodDelete, path, handler)
}

// UseMetal registers an http package based middleware that is run before each request
func (r *Router[ReqCtx]) UseMetal(fn func(http.ResponseWriter, *http.Request, http.Handler)) {
	r.metal = append(r.metal, fn)
}

// Use registers a middleware that will be run after the UseMetal middleware but
// before the handler. Each middleware is passed the next Handler or middleware
// in the stack. Not calling the next function will halt the middleware/handler chain.
func (r *Router[ReqCtx]) Use(fn func(context.Context, ReqCtx, Handler[ReqCtx])) {
	r.middleware = append(r.middleware, fn)
}

// Group returns a new route group with the given prefix. The group can define
// its own middleware that will only be run for that group.
func (r *Router[ReqCtx]) Group(prefix string) *Group[ReqCtx] {
	return NewGroup[ReqCtx](r, prefix)
}

// ServeHTTP implements the http.Handler interface.
func (r *Router[ReqCtx]) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Run fernet middleware and call route handler
	handler := func(rw http.ResponseWriter, req *http.Request) {
		method := req.Method
		normalizedPath := normalizeRoutePath(req.URL.Path)
		lookup := []string{method}
		lookup = append(lookup, normalizedPath...)

		ok, value := r.tree.Value(lookup)
		if !ok {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

		ok, params := value.match(req)
		if !ok {
			// This should never actually get hit in real code but would
			// indicate a bug in the framework.
			panic("route did not match request")
		}

		handler := func(ctx context.Context, rctx ReqCtx) {
			value.handler(ctx, rctx)
		}

		for i := len(r.middleware) - 1; i >= 0; i-- {
			currentHandler := handler
			middleware := r.middleware[i]
			handler = func(ctx context.Context, reqCtx ReqCtx) {
				middleware(ctx, reqCtx, currentHandler)
			}
		}

		handler(
			req.Context(),
			r.initReqCtx(newRequestContext(req, rw, value.Path, params)),
		)
	}

	// Run Metal middleware
	for i := len(r.metal) - 1; i >= 0; i-- {
		currentHandler := handler
		metal := r.metal[i]

		handler = func(rw http.ResponseWriter, r *http.Request) {
			metal(rw, r, http.HandlerFunc(currentHandler))
		}
	}

	handler(rw, req)
}

type Registerable[T RequestContext] interface {
	Register(Routable[T])
}

// Register passes this routable to the provided registerable so that it can
// register its own routes on the routable. This is useful for building
// abstractions like controllers or packages that need to manage and register
// their own routes/state.
func (r *Router[ReqCtx]) Register(c Registerable[ReqCtx]) {
	c.Register(r)
}
