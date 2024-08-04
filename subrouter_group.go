package fernet

import (
	"context"
	"net/http"
	"reflect"
)

// controllerGroup is a group of routes from a controller that share a common
// prefix.
type controllerGroup[T RequestContext, RequestData FromRequest[T]] struct {
	prefix      string
	parent      Registerable[T]
	middlewares []func(context.Context, T, Handler[T])
}

var _ ControllerRoutable[*RootRequestContext, *placeholderFromRequest] = &Controller[*RootRequestContext, *placeholderFromRequest]{}

// RawMatch implements the Registerable interface and forwards the call to the
// parent router. This allows other controllers and controller groups to be
// registered with the controller.
func (r *controllerGroup[T, RequestData]) RawMatch(method string, path string, fn Handler[T]) {
	r.parent.RawMatch(method, joinURL(r.prefix, path), r.wrap(fn))
}

// Match registers the given handler with the given method and path.
func (r *controllerGroup[T, RequestData]) Match(method string, path string, fn ControllerHandler[T, RequestData]) {
	r.parent.RawMatch(method, joinURL(r.prefix, path), r.wrap(r.normalizeHandler(fn)))
}

// Get registers a GET handler with the given path.
func (r *controllerGroup[T, RequestData]) Get(path string, fn ControllerHandler[T, RequestData]) {
	r.Match(http.MethodGet, path, fn)
}

// Post registers a POST handler with the given path.
func (r *controllerGroup[T, RequestData]) Post(path string, fn ControllerHandler[T, RequestData]) {
	r.Match(http.MethodPost, path, fn)
}

// Put registers a PUT handler with the given path.
func (r *controllerGroup[T, RequestData]) Put(path string, fn ControllerHandler[T, RequestData]) {
	r.Match(http.MethodPut, path, fn)
}

// Patch registers a PATCH handler with the given path.
func (r *controllerGroup[T, RequestData]) Patch(path string, fn ControllerHandler[T, RequestData]) {
	r.Match(http.MethodPatch, path, fn)
}

// Delete registers a DELETE handler with the given path.
func (r *controllerGroup[T, RequestData]) Delete(path string, fn ControllerHandler[T, RequestData]) {
	r.Match(http.MethodDelete, path, fn)
}

// Group returns a new controller group with the given prefix.
func (r *controllerGroup[T, RequestData]) Group() *controllerGroup[T, RequestData] {
	return &controllerGroup[T, RequestData]{
		parent: r,
	}
}

// Namespace returns a new controller group with the given prefix.
func (r *controllerGroup[T, RequestData]) Namespace(prefix string) *controllerGroup[T, RequestData] {
	return &controllerGroup[T, RequestData]{
		prefix: prefix,
		parent: r,
	}
}

// Use registers a middleware function that will be called before each handler.
// Middleware are always called before FromRequest.
func (r *controllerGroup[T, RequestData]) Use(fn func(context.Context, T, Handler[T])) {
	r.middlewares = append(r.middlewares, fn)
}

func (r *controllerGroup[T, RequestData]) wrap(fn Handler[T]) Handler[T] {
	handler := fn

	for _, middleware := range r.middlewares {
		currentHandler := handler
		handler = func(ctx context.Context, rc T) {
			middleware(ctx, rc, currentHandler)
		}
	}

	return handler
}

func (r *controllerGroup[T, RequestData]) normalizeHandler(fn ControllerHandler[T, RequestData]) Handler[T] {
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
