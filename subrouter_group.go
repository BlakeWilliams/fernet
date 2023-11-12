package fernet

import (
	"context"
	"net/http"
	"reflect"
	"strings"
)

// subRouterGroup is a group of routes from a SubRouter that share a common
// prefix.
type subRouterGroup[T RequestContext, RequestData FromRequest[T]] struct {
	prefix  string
	parent  Registerable[T]
	befores []func(context.Context, T, RequestData) bool
}

var _ SubRouterRoutable[*RootRequestContext, *placeholderFromRequest] = &SubRouter[*RootRequestContext, *placeholderFromRequest]{}

// RawMatch implements the Registerable interface and forwards the call to the
// parent router. This allows subrouters and groups to be registered with the
// subrouter.
func (r *subRouterGroup[T, RequestData]) RawMatch(method string, path string, fn Handler[T]) {
	path = strings.TrimSuffix(r.prefix, "/") + "/" + strings.TrimPrefix(path, "/")
	r.parent.RawMatch(method, path, fn)
}

// Match registers the given handler with the given method and path.
func (r *subRouterGroup[T, RequestData]) Match(method string, path string, fn SubRouterHandler[T, RequestData]) {
	path = strings.TrimSuffix(r.prefix, "/") + "/" + strings.TrimPrefix(path, "/")
	r.parent.RawMatch(method, path, r.wrap(fn))
}

// Get registers a GET handler with the given path.
func (r *subRouterGroup[T, RequestData]) Get(path string, fn SubRouterHandler[T, RequestData]) {
	r.Match(http.MethodGet, path, fn)
}

// Post registers a POST handler with the given path.
func (r *subRouterGroup[T, RequestData]) Post(path string, fn SubRouterHandler[T, RequestData]) {
	r.Match(http.MethodPost, path, fn)
}

// Put registers a PUT handler with the given path.
func (r *subRouterGroup[T, RequestData]) Put(path string, fn SubRouterHandler[T, RequestData]) {
	r.Match(http.MethodPut, path, fn)
}

// Patch registers a PATCH handler with the given path.
func (r *subRouterGroup[T, RequestData]) Patch(path string, fn SubRouterHandler[T, RequestData]) {
	r.Match(http.MethodPatch, path, fn)
}

// Delete registers a DELETE handler with the given path.
func (r *subRouterGroup[T, RequestData]) Delete(path string, fn SubRouterHandler[T, RequestData]) {
	r.Match(http.MethodDelete, path, fn)
}

// Group returns a new SubRouterGroup with the given prefix.
func (r *subRouterGroup[T, RequestData]) Group(prefix string) *subRouterGroup[T, RequestData] {
	return &subRouterGroup[T, RequestData]{
		prefix: prefix,
		parent: r,
	}
}

func (r *subRouterGroup[T, RequestData]) wrap(fn SubRouterHandler[T, RequestData]) Handler[T] {
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

		fn(ctx, rc, requestData.(RequestData))
	}
}
