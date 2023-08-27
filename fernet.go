package fernet

import (
	"net/http"

	"github.com/blakewilliams/fernet/internal/radical"
)

type (
	// Handler is a function that accepts a ResponseWriter and a Request.
	Handler[RequestData any] func(Response, *Request[RequestData])

	// Router represents the primary router for the application.
	Router[RequestData any] struct {
		routes     []*route[RequestData]
		tree       *radical.Node[*route[RequestData]]
		metal      []func(http.ResponseWriter, *http.Request, http.Handler)
		middleware []func(Response, *Request[RequestData], Handler[RequestData])
	}

	// Routable is an interface that can be implemented by types that want to
	// register routes with a router.
	Routable[RequestData any] interface {
		// Match registers a route with the given method and path
		Match(method string, path string, fn Handler[RequestData])
		// Get registers a GET route with the given path
		Get(method string, fn Handler[RequestData])
		// Post registers a POST route with the given path
		Post(method string, fn Handler[RequestData])
		// Put registers a PUT route with the given path
		Put(method string, fn Handler[RequestData])
		// Patch registers a PATCH route with the given path
		Patch(method string, fn Handler[RequestData])
		// Delete registers a DELETE route with the given path
		Delete(method string, fn Handler[RequestData])

		// Use registers a middleware function that is run before each request
		// for this group and all groups below it.
		Use(func(Response, *Request[RequestData], Handler[RequestData]))

		// Group returns a new group based on this Routable. It will have its
		// own middleware stack in addition to the middleware stack on the
		// groups/router above it.
		Group(prefix string) *Group[RequestData]
	}
)

var _ Routable[int] = (*Router[int])(nil)

// New returns a new router instance. It accepts a generic type for RequestData,
// which is passed to each middleware and handler so that they can share data
// throughout the routing stack.
//
// This is useful for sharing request specific data like the authenticated user,
// or other data needed by the controller, handler, template, or other layers.
func New[RequestData any]() *Router[RequestData] {
	return &Router[RequestData]{
		tree:       radical.New[*route[RequestData]](),
		routes:     make([]*route[RequestData], 0),
		metal:      make([]func(http.ResponseWriter, *http.Request, http.Handler), 0),
		middleware: make([]func(Response, *Request[RequestData], Handler[RequestData]), 0),
	}
}

// Match registers a route with the router.
func (r *Router[RequestData]) Match(method string, path string, handler Handler[RequestData]) {
	route := newRoute[RequestData](method, path, handler)
	r.routes = append(r.routes, route)

	pathParts := make([]string, 0, len(route.parts)+1)
	pathParts = append(pathParts, method)
	pathParts = append(pathParts, route.parts...)

	r.tree.Add(pathParts, route)
}

// Get registers a GET route with the router.
func (r *Router[RequestData]) Get(path string, handler Handler[RequestData]) {
	r.Match(http.MethodGet, path, handler)
}

// Get registers a GET route with the router.
func (r *Router[RequestData]) Post(path string, handler Handler[RequestData]) {
	r.Match(http.MethodPost, path, handler)
}

// Put registers a PUT route with the router.
func (r *Router[RequestData]) Put(path string, handler Handler[RequestData]) {
	r.Match(http.MethodPut, path, handler)
}

// Patch registers a PATCH route with the router.
func (r *Router[RequestData]) Patch(path string, handler Handler[RequestData]) {
	r.Match(http.MethodPatch, path, handler)
}

// Delete registers a DELETE route with the router.
func (r *Router[RequestData]) Delete(path string, handler Handler[RequestData]) {
	r.Match(http.MethodDelete, path, handler)
}

// UseMetal registers an http package based middleware that is run before each request
func (r *Router[RequestData]) UseMetal(fn func(http.ResponseWriter, *http.Request, http.Handler)) {
	r.metal = append(r.metal, fn)
}

// Use registers a middleware that will be run after the UseMetal middleware but
// before the handler. Each middleware is passed the next Handler or middleware
// in the stack. Not calling the next function will halt the middleware/handler chain.
func (r *Router[RequestData]) Use(fn func(Response, *Request[RequestData], Handler[RequestData])) {
	r.middleware = append(r.middleware, fn)
}

// Group returns a new route group with the given prefix. The group can define
// its own middleware that will only be run for that group.
func (r *Router[RequestData]) Group(prefix string) *Group[RequestData] {
	return NewGroup[RequestData](r, prefix)
}

// ServeHTTP implements the http.Handler interface.
func (r *Router[RequestData]) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
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

		ok, params := value.Match(req)
		if !ok {
			// This should never actually get hit in real code but would
			// indicate a bug in the framework.
			panic("route did not match request")
		}

		request := &Request[RequestData]{req: req, params: params}
		response := &response{header: make(http.Header)}

		handler := value.handler

		for i := len(r.middleware) - 1; i >= 0; i-- {
			currentHandler := handler
			middleware := r.middleware[i]

			handler = func(res Response, req *Request[RequestData]) {
				middleware(res, req, currentHandler)
			}
		}

		handler(response, request)

		for k, v := range response.header {
			rw.Header()[k] = v
		}

		if response.status == 0 {
			rw.WriteHeader(http.StatusOK)
		} else {
			rw.WriteHeader(response.status)
		}

		_, _ = rw.Write(response.body)
	}

	for i := len(r.metal) - 1; i >= 0; i-- {
		currentHandler := handler
		metal := r.metal[i]

		handler = func(rw http.ResponseWriter, r *http.Request) {
			metal(rw, r, http.HandlerFunc(currentHandler))
		}
	}

	handler(rw, req)
}

type Registerable[T any] interface {
	Register(Routable[T])
}

// Register passes this routable to the provided registerable so that it can
// register its own routes on the routable. This is useful for building
// abstractions like controllers or packages that need to manage and register
// their own routes/state.
func (r *Router[RequestData]) Register(c Registerable[RequestData]) {
	c.Register(r)
}
