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
)

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

func (r *Router[RequestData]) Use(fn func(Response, *Request[RequestData], Handler[RequestData])) {
	r.middleware = append(r.middleware, fn)
}

// ServeHTTP implements the http.Handler interface.
func (r *Router[RequestData]) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Run fernet middleware and call route handler
	handler := func(rw http.ResponseWriter, req *http.Request) {
		request := &Request[RequestData]{req: req}
		response := &response{header: make(http.Header)}

		method := request.Method()
		normalizedPath := normalizeRoutePath(request.URL().Path)
		lookup := []string{method}
		lookup = append(lookup, normalizedPath...)

		ok, value := r.tree.Value(lookup)
		if !ok {
			rw.WriteHeader(http.StatusNotFound)
			return
		}

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
