package fernet

import (
	"context"
	"net/http"
	"net/url"
)

// Request represents an incoming HTTP request.
type Request[Data any] struct {
	req    *http.Request
	ctx    context.Context
	params map[string]string
	Data   Data
}

// Context returns the context associated with this Request. It returns the
// original *http.Request if no context was explicitly set.
func (r *Request[Data]) Context() context.Context {
	if r.ctx != nil {
		return r.ctx
	}

	return r.req.Context()
}

// WithContext returns a copy of this request with a new context.
func (r *Request[Data]) WithContext(ctx context.Context) *Request[Data] {
	return &Request[Data]{
		req:    r.req,
		ctx:    ctx,
		params: r.params,
		Data:   r.Data,
	}
}

// Method returns the HTTP method of the request.
func (r *Request[Data]) Method() string {
	return r.req.Method
}

// URL returns the URL of the request.
func (r *Request[Data]) URL() url.URL {
	return *r.req.URL
}

// Param returns the value for a matching route parameter. It returns an empty
// string if the value does not exist.
//
// e.g. req.Param("name") returns "Fox" for: /hello/:foo => /hello/fox
func (r *Request[Data]) Param(name string) string {
	return r.params[name]
}

// Param returns the matching route parameters.
//
// e.g. /greet/:name/:greeting => /greet/fox/hello results in: map[string]string{"name": Fox", "greeting": "hello"}
func (r *Request[Data]) Params() map[string]string {
	return r.params
}
