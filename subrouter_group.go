package fernet

import (
	"context"
	"net/http"
	"reflect"
)

// SubRouterGroup is a group of routes from a SubRouter that share a common
// prefix and maintain their own befores stack.
type SubRouterGroup[T RequestContext, RequestData FromRequest[T]] struct {
	prefix  string
	parent  Registerable[T]
	befores []func(context.Context, T, RequestData) bool
}

var _ SubRouterRoutable[*RootRequestContext, *placeholderFromRequest] = &SubRouter[*RootRequestContext, *placeholderFromRequest]{}

// RawMatch implements the Registerable interface and forwards the call to the
// parent router. This allows subrouters and groups to be registered with the
// subrouter.
func (r *SubRouterGroup[T, RequestData]) RawMatch(method string, path string, fn Handler[T]) {
	r.parent.RawMatch(method, path, fn)
}

// Match registers the given handler with the given method and path.
func (r *SubRouterGroup[T, RequestData]) Match(method string, path string, fn SubRouterHandler[T, RequestData]) {
	r.parent.RawMatch(method, path, r.wrap(fn))
}

// Get registers a GET handler with the given path.
func (r *SubRouterGroup[T, RequestData]) Get(path string, fn SubRouterHandler[T, RequestData]) {
	r.Match(http.MethodGet, path, fn)
}

// Post registers a POST handler with the given path.
func (r *SubRouterGroup[T, RequestData]) Post(path string, fn SubRouterHandler[T, RequestData]) {
	r.Match(http.MethodPost, path, fn)
}

// Put registers a PUT handler with the given path.
func (r *SubRouterGroup[T, RequestData]) Put(path string, fn SubRouterHandler[T, RequestData]) {
	r.Match(http.MethodPut, path, fn)
}

// Patch registers a PATCH handler with the given path.
func (r *SubRouterGroup[T, RequestData]) Patch(path string, fn SubRouterHandler[T, RequestData]) {
	r.Match(http.MethodPatch, path, fn)
}

// Delete registers a DELETE handler with the given path.
func (r *SubRouterGroup[T, RequestData]) Delete(path string, fn SubRouterHandler[T, RequestData]) {
	r.Match(http.MethodDelete, path, fn)
}

// Before registers a befores function that will be called before each handler.
// If the function returns false, the subsequent befores and the handler will
// not be called.
func (r *SubRouterGroup[T, RequestData]) Before(fn func(context.Context, T, RequestData) bool) {
	r.befores = append(r.befores, fn)
}

// Group returns a new SubRouterGroup with the given prefix.
func (r *SubRouterGroup[T, RequestData]) Group(prefix string) *SubRouterGroup[T, RequestData] {
	return &SubRouterGroup[T, RequestData]{
		prefix:  r.prefix + prefix,
		parent:  r,
		befores: make([]func(context.Context, T, RequestData) bool, 0),
	}
}

func (r *SubRouterGroup[T, RequestData]) wrap(fn SubRouterHandler[T, RequestData]) Handler[T] {
	var t RequestData
	requestDataType := reflect.TypeOf(t)
	isPointer := requestDataType.Kind() == reflect.Ptr

	if isPointer {
		requestDataType = requestDataType.Elem()
	}

	return func(ctx context.Context, rc T) {
		newRequestData := reflect.New(requestDataType)
		if !isPointer {
			newRequestData = newRequestData.Elem()
		}
		requestData := newRequestData.Interface()
		success := (requestData).(FromRequest[T]).FromRequest(ctx, rc)

		if !success {
			return
		}

		for _, before := range r.befores {
			if !before(ctx, rc, requestData.(RequestData)) {
				return
			}
		}

		fn(ctx, rc, requestData.(RequestData))
	}
}
