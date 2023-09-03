package fernet

import "net/http"

type (
	// Metal is an http middleware that is run before fernet handlers and middleware
	Metal func(http.ResponseWriter, *http.Request, http.Handler)

	// MetalStack is a stack of metal middleware. Metal are called in order of
	// registration and can halt the execution of the request.
	MetalStack[T RequestContext] struct {
		stack []Metal
		entry http.Handler
	}
)

// NewMetalStack returns a new http middleware stack that will can register
// middleware that will be called before the provided entry http.Handler
func NewMetalStack[T RequestContext](entry http.Handler) *MetalStack[T] {
	return &MetalStack[T]{
		stack: make([]Metal, 0),
		entry: entry,
	}
}

// Use registers the provided function as a middleware that will be called
// before fernet handlers and middleware.
func (ms *MetalStack[T]) Use(fn Metal) {
	ms.stack = append(ms.stack, fn)
}

// ServeHTTP runs each of the middleware, passing the next middleware in the
// stack as a callback. The final middleware called is always the `entry`
// http.Handler passed to `NewMetalStack.
func (ms *MetalStack[T]) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	handler := ms.entry

	for i := len(ms.stack) - 1; i >= 0; i-- {
		currentHandler := handler
		middleware := ms.stack[i]

		handler = http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			middleware(rw, r, currentHandler)
		})
	}

	handler.ServeHTTP(rw, r)
}
