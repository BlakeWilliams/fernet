package fernet

import (
	"net/http"
)

// RequestContext is an interface that exposes the http.Request,
// http.ResponseWriter, and route params to a handler. Custom types can
// implement this interface to be passed to handlers of the router.
type RequestContext interface {
	// Request returns the original *http.Request
	Request() *http.Request
	// Writer returns a fernet.Response
	Response() Response
	// Params returns the parameters extracted from the URL path based on the
	// matched route.
	Params() map[string]string
	// MatchedPath returns the path that was matched by the router.
	MatchedPath() string
}

// BasicRequestContext is a basic implementation of RequestContext. It can be embedded in
// other types to provide a default implementation of the RequestContext interface.
type RootRequestContext struct {
	req         *http.Request
	res         Response
	params      map[string]string
	matchedPath string
}

var _ RequestContext = (*RootRequestContext)(nil)

func NewRequestContext(req *http.Request, res http.ResponseWriter, matchedPath string, params map[string]string) *RootRequestContext {
	return &RootRequestContext{
		req:         req,
		res:         newResponseWriter(res),
		matchedPath: matchedPath,
		params:      params,
	}
}

func (r *RootRequestContext) Request() *http.Request {
	return r.req
}

func (r *RootRequestContext) Response() Response {
	return r.res
}

func (r *RootRequestContext) Params() map[string]string {
	return r.params
}

func (r *RootRequestContext) MatchedPath() string {
	return r.matchedPath
}
