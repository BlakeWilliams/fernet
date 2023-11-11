package fernet

import (
	"context"
	"net/http"
	"reflect"
)

// SubRouterGroup is a group of routes from a SubRouter that share a common
// prefix and maintain their own middleware stack.
type SubRouterGroup[T RequestContext, Child FromRequest[T]] struct {
	prefix     string
	parent     Registerable[T]
	middleware []func(context.Context, T, Handler[T])
}

var _ SubRouterRoutable[*RootRequestContext, *placeholderFromRequest] = &SubRouter[*RootRequestContext, *placeholderFromRequest]{}

// RawMatch implements the Registerable interface and forwards the call to the
// parent router. This allows subrouters and groups to be registered with the
// subrouter.
func (r *SubRouterGroup[T, Child]) RawMatch(method string, path string, fn Handler[T]) {
	r.parent.RawMatch(method, path, fn)
}

// Match registers the given handler with the given method and path.
func (r *SubRouterGroup[T, Child]) Match(method string, path string, fn SubRouterHandler[T, Child]) {
	r.parent.RawMatch(method, path, r.wrap(fn))
}

// Get registers a GET handler with the given path.
func (r *SubRouterGroup[T, Child]) Get(path string, fn SubRouterHandler[T, Child]) {
	r.Match(http.MethodGet, path, fn)
}

// Post registers a POST handler with the given path.
func (r *SubRouterGroup[T, Child]) Post(path string, fn SubRouterHandler[T, Child]) {
	r.Match(http.MethodPost, path, fn)
}

// Put registers a PUT handler with the given path.
func (r *SubRouterGroup[T, Child]) Put(path string, fn SubRouterHandler[T, Child]) {
	r.Match(http.MethodPut, path, fn)
}

// Patch registers a PATCH handler with the given path.
func (r *SubRouterGroup[T, Child]) Patch(path string, fn SubRouterHandler[T, Child]) {
	r.Match(http.MethodPatch, path, fn)
}

// Delete registers a DELETE handler with the given path.
func (r *SubRouterGroup[T, Child]) Delete(path string, fn SubRouterHandler[T, Child]) {
	r.Match(http.MethodDelete, path, fn)
}

// Use registers a middleware function that will be called before each handler.
func (r *SubRouterGroup[T, Child]) Use(fn func(context.Context, T, Handler[T])) {
	r.middleware = append(r.middleware, fn)
}

// Group returns a new SubRouterGroup with the given prefix.
func (r *SubRouterGroup[T, Child]) Group(prefix string) *SubRouterGroup[T, Child] {
	return &SubRouterGroup[T, Child]{
		prefix:     r.prefix + prefix,
		parent:     r,
		middleware: make([]func(context.Context, T, Handler[T]), 0),
	}
}

func (r *SubRouterGroup[T, Child]) wrap(fn SubRouterHandler[T, Child]) Handler[T] {
	var t Child
	childType := reflect.TypeOf(t)
	isPointer := childType.Kind() == reflect.Ptr

	if isPointer {
		childType = childType.Elem()
	}

	return func(ctx context.Context, rc T) {
		newChild := reflect.New(childType)
		if !isPointer {
			// TODO report warning?
			newChild = newChild.Elem()
		}
		child := newChild.Interface()
		success := (child).(FromRequest[T]).FromRequest(ctx, rc)

		if !success {
			return
		}

		fn(ctx, rc, child.(Child))
	}
}
