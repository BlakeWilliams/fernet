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

	// Controller is similar to fernet.Router, but accepts a type that implements
	// the FromRequest interface that will be initialized each request and passed
	// to the handler as the third argument.
	Controller[T RequestContext, RequestData FromRequest[T]] struct {
		parent Registerable[T]
		root   *controllerGroup[T, RequestData]
	}

	// ControllerHandler is the signature for controller handlers. It accepts the
	// standard context.Context, and T RequestContext, but also a third argument
	// that implements the FromRequest interface.
	ControllerHandler[T RequestContext, RequestData FromRequest[T]] func(context.Context, T, RequestData)

	// ControllerRoutable ensures consistency across all controller based types.
	ControllerRoutable[T RequestContext, RequestData FromRequest[T]] interface {
		Match(string, string, ControllerHandler[T, RequestData])
		Get(string, ControllerHandler[T, RequestData])
		Post(string, ControllerHandler[T, RequestData])
		Put(string, ControllerHandler[T, RequestData])
		Patch(string, ControllerHandler[T, RequestData])
		Delete(string, ControllerHandler[T, RequestData])
		Use(...func(context.Context, T, Handler[T]))
	}

	placeholderFromRequest struct{}
)

// Implement the FromRequest interface for the placeholder type so we can assert interface adherence.
func (p *placeholderFromRequest) FromRequest(context.Context, *RootRequestContext) bool { return false }

var _ ControllerRoutable[*RootRequestContext, *placeholderFromRequest] = &Controller[*RootRequestContext, *placeholderFromRequest]{}

// NewController creates a new controller that can be used to register handlers
// that accept a type that implements the FromRequest interface. Each request
// will initialize a new instance of the type, call `FromRequest` on it, and
// pass it to the handler if the method returns true.
func NewController[Parent RequestContext, RequestData FromRequest[Parent]](r Registerable[Parent], dataType RequestData) *Controller[Parent, RequestData] {
	return &Controller[Parent, RequestData]{
		parent: r,
		root: &controllerGroup[Parent, RequestData]{
			prefix:      "",
			parent:      r,
			middlewares: make([]func(context.Context, Parent, Handler[Parent]), 0),
		},
	}
}

// RawMatch implements the Registerable interface and forwards the call to the
// parent router. This allows controllers and groups to be registered with the
// current controller.
func (r *Controller[T, RequestData]) RawMatch(method string, path string, fn Handler[T]) {
	r.parent.RawMatch(method, path, fn)
}

// Match registers the given handler with the given method and path.
func (r *Controller[T, RequestData]) Match(method string, path string, fn ControllerHandler[T, RequestData]) {
	r.root.Match(method, path, fn)
}

// Get registers a GET handler with the given path.
func (r *Controller[T, RequestData]) Get(path string, fn ControllerHandler[T, RequestData]) {
	r.root.Get(path, fn)
}

// Post registers a POST handler with the given path.
func (r *Controller[T, RequestData]) Post(path string, fn ControllerHandler[T, RequestData]) {
	r.root.Post(path, fn)
}

// Put registers a PUT handler with the given path.
func (r *Controller[T, RequestData]) Put(path string, fn ControllerHandler[T, RequestData]) {
	r.root.Put(path, fn)
}

// Patch registers a PATCH handler with the given path.
func (r *Controller[T, RequestData]) Patch(path string, fn ControllerHandler[T, RequestData]) {
	r.root.Patch(path, fn)
}

// Delete registers a DELETE handler with the given path.
func (r *Controller[T, RequestData]) Delete(path string, fn ControllerHandler[T, RequestData]) {
	r.root.Delete(path, fn)
}

// Group returns a new ControllerGroup.
func (r *Controller[T, RequestData]) Group() *controllerGroup[T, RequestData] {
	return r.root.Group()
}

// Namespace returns a new ControllerGroup with the given prefix applied to all routes defined on it.
func (r *Controller[T, RequestData]) Namespace(prefix string) *controllerGroup[T, RequestData] {
	return r.root.Namespace(prefix)
}

// Use registers a middleware function that will be called before each request.
// Middlewares are always called in the order they are registered and before
// FromRequest is called.
func (r *Controller[T, RequestData]) Use(fns ...func(context.Context, T, Handler[T])) {
	r.root.Use(fns...)
}
