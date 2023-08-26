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
		routes []*route[RequestData]
		tree   *radical.Node[*route[RequestData]]
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
		tree:   radical.New[*route[RequestData]](),
		routes: make([]*route[RequestData], 0),
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

// ServeHTTP implements the http.Handler interface.
func (r *Router[RequestData]) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	request := &Request[RequestData]{req: req}
	response := &response{header: make(http.Header)}

	method := request.Method()
	normalizedPath := normalizeRoutePath(request.URL().Path)
	lookup := []string{method}
	lookup = append(lookup, normalizedPath...)

	ok, value := r.tree.Value(lookup)
	if !ok {
		res.WriteHeader(http.StatusNotFound)
		return
	}

	value.handler(response, request)

	for k, v := range response.header {
		res.Header()[k] = v
	}

	res.WriteHeader(response.status)
	_, _ = res.Write(response.body)
}
