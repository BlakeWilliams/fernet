package fernet

import (
	"context"
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
	SubRouter[T RequestContext, RequestData FromRequest[T]] struct {
		parent Registerable[T]
		root   *SubRouterGroup[T, RequestData]
	}

	// SubRouterHandler is the signature for SubRouter handlers. It accepts the
	// standard context.Context, and T RequestContext, but also a third argument
	// that implements the FromRequest interface.
	SubRouterHandler[T RequestContext, RequestData FromRequest[T]] func(context.Context, T, RequestData)

	// SubRouterRoutable ensures consistency across all SubRouter based types.
	SubRouterRoutable[T RequestContext, RequestData FromRequest[T]] interface {
		Match(string, string, SubRouterHandler[T, RequestData])
		Get(string, SubRouterHandler[T, RequestData])
		Post(string, SubRouterHandler[T, RequestData])
		Put(string, SubRouterHandler[T, RequestData])
		Patch(string, SubRouterHandler[T, RequestData])
		Delete(string, SubRouterHandler[T, RequestData])
		Before(func(context.Context, T, RequestData) bool)
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
func NewSubRouter[Parent RequestContext, RequestData FromRequest[Parent]](r Registerable[Parent], dataType RequestData) *SubRouter[Parent, RequestData] {
	return &SubRouter[Parent, RequestData]{
		parent: r,
		root: &SubRouterGroup[Parent, RequestData]{
			prefix:  "",
			parent:  r,
			befores: make([]func(context.Context, Parent, RequestData) bool, 0),
		},
	}
}

// RawMatch implements the Registerable interface and forwards the call to the
// parent router. This allows subrouters and groups to be registered with the
// subrouter.
func (r *SubRouter[T, RequestData]) RawMatch(method string, path string, fn Handler[T]) {
	r.parent.RawMatch(method, path, fn)
}

// Match registers the given handler with the given method and path.
func (r *SubRouter[T, RequestData]) Match(method string, path string, fn SubRouterHandler[T, RequestData]) {
	r.root.Match(method, path, fn)
}

// Get registers a GET handler with the given path.
func (r *SubRouter[T, RequestData]) Get(path string, fn SubRouterHandler[T, RequestData]) {
	r.root.Get(path, fn)
}

// Post registers a POST handler with the given path.
func (r *SubRouter[T, RequestData]) Post(path string, fn SubRouterHandler[T, RequestData]) {
	r.root.Post(path, fn)
}

// Put registers a PUT handler with the given path.
func (r *SubRouter[T, RequestData]) Put(path string, fn SubRouterHandler[T, RequestData]) {
	r.root.Put(path, fn)
}

// Patch registers a PATCH handler with the given path.
func (r *SubRouter[T, RequestData]) Patch(path string, fn SubRouterHandler[T, RequestData]) {
	r.root.Patch(path, fn)
}

// Delete registers a DELETE handler with the given path.
func (r *SubRouter[T, RequestData]) Delete(path string, fn SubRouterHandler[T, RequestData]) {
	r.root.Delete(path, fn)
}

// Use registers a middleware function that will be called before each handler.
func (r *SubRouter[T, RequestData]) Before(fn func(context.Context, T, RequestData) bool) {
	r.root.Before(fn)
}

// Group returns a new SubRouterGroup with the given prefix.
func (r *SubRouter[T, RequestData]) Group(prefix string) *SubRouterGroup[T, RequestData] {
	return r.root.Group(prefix)
}
