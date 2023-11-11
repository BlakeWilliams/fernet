package fernet

import (
	"context"
	"net/http"
)

type (
	// FromRequest enables a struct to be initialized from a RequestContext that
	// will be passed to each handler. It accepts a context.Context and the
	// generic RequestContext type.
	//
	// The request can be short-circuited by returning false from this function.
	FromRequest[T RequestContext] interface {
		FromRequest(context.Context, T) bool
	}

	// SubRouter is similar to fernet.Router, but accepts a type that implements
	// the FromRequest interface that will be initialized each request and passed
	// to the handler as the third argument.
	SubRouter[T RequestContext, Child FromRequest[T]] struct {
		parent     Registerable[T]
		root       *SubRouterGroup[T, Child]
		middleware []func(context.Context, T, Handler[T])
	}

	// SubRouterHandler is the signature for SubRouter handlers. It accepts the
	// standard context.Context, and T RequestContext, but also a third argument
	// that implements the FromRequest interface.
	SubRouterHandler[T RequestContext, Child FromRequest[T]] func(context.Context, T, Child)

	// SubRouterRoutable ensures consistency across all SubRouter based types.
	SubRouterRoutable[T RequestContext, Child FromRequest[T]] interface {
		Match(string, string, SubRouterHandler[T, Child])
		Get(string, SubRouterHandler[T, Child])
		Post(string, SubRouterHandler[T, Child])
		Put(string, SubRouterHandler[T, Child])
		Patch(string, SubRouterHandler[T, Child])
		Delete(string, SubRouterHandler[T, Child])

		Use(func(context.Context, T, Handler[T]))
		Group(string) *SubRouterGroup[T, Child]
	}

	placeholderFromRequest struct{}
)

// Implement the FromRequest interface for the placeholder type so we can assert interface adherence.
func (p *placeholderFromRequest) FromRequest(context.Context, *RootRequestContext) bool { return false }

var _ SubRouterRoutable[*RootRequestContext, *placeholderFromRequest] = &SubRouter[*RootRequestContext, *placeholderFromRequest]{}

// NewSubRouter creates a new subrouter that can be used to register handlers
// that accept a type that implements the FromRequest interface. Each request
// will initialize a new instance of the type, call `FromRequest` on it, and
// pass it to the handler if the method returns true.
func NewSubRouter[Parent RequestContext, Child FromRequest[Parent]](r Registerable[Parent], dataType Child) *SubRouter[Parent, Child] {
	return &SubRouter[Parent, Child]{
		parent: r,
		root: &SubRouterGroup[Parent, Child]{
			prefix:     "",
			parent:     r,
			middleware: make([]func(context.Context, Parent, Handler[Parent]), 0),
		},
		middleware: make([]func(context.Context, Parent, Handler[Parent]), 0),
	}
}

// RawMatch implements the Registerable interface and forwards the call to the
// parent router. This allows subrouters and groups to be registered with the
// subrouter.
func (r *SubRouter[T, Child]) RawMatch(method string, path string, fn Handler[T]) {
	r.parent.RawMatch(method, path, fn)
}

// Match registers the given handler with the given method and path.
func (r *SubRouter[T, Child]) Match(method string, path string, fn SubRouterHandler[T, Child]) {
	r.root.Match(method, path, fn)
}

// Get registers a GET handler with the given path.
func (r *SubRouter[T, Child]) Get(path string, fn SubRouterHandler[T, Child]) {
	r.Match(http.MethodGet, path, fn)
}

// Post registers a POST handler with the given path.
func (r *SubRouter[T, Child]) Post(path string, fn SubRouterHandler[T, Child]) {
	r.Match(http.MethodPost, path, fn)
}

// Put registers a PUT handler with the given path.
func (r *SubRouter[T, Child]) Put(path string, fn SubRouterHandler[T, Child]) {
	r.Match(http.MethodPut, path, fn)
}

// Patch registers a PATCH handler with the given path.
func (r *SubRouter[T, Child]) Patch(path string, fn SubRouterHandler[T, Child]) {
	r.Match(http.MethodPatch, path, fn)
}

// Delete registers a DELETE handler with the given path.
func (r *SubRouter[T, Child]) Delete(path string, fn SubRouterHandler[T, Child]) {
	r.Match(http.MethodDelete, path, fn)
}

// Use registers a middleware function that will be called before each handler.
func (r *SubRouter[T, Child]) Use(fn func(context.Context, T, Handler[T])) {
	r.middleware = append(r.middleware, fn)
}

// Group returns a new SubRouterGroup with the given prefix.
func (r *SubRouter[T, Child]) Group(prefix string) *SubRouterGroup[T, Child] {
	return r.root.Group(prefix)
}
