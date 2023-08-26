package fernet

import (
	"net/http"
	"net/url"
)

// Request represents an incoming HTTP request.
type Request[Data any] struct {
	req *http.Request
}

// Method returns the HTTP method of the request.
func (r *Request[Data]) Method() string {
	return r.req.Method
}

// URL returns the URL of the request.
func (r *Request[Data]) URL() url.URL {
	return *r.req.URL
}
