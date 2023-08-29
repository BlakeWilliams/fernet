package fernet

import (
	"context"
	"net/http"
)

type Request interface {
	// Request returns the original *http.Request
	Request() *http.Request
	Params() map[string]string
}

type Response interface {
	// Writer returns the original http.ResponseWriter.
	ResponseWriter() http.ResponseWriter
}

type ReqRes interface {
	Request
	Response
}

type RequestContext interface {
	// Request returns the original *http.Request
	Request() *http.Request
	// Writer returns the original http.ResponseWriter. This can conflict with
	// the ResponseWriter interface, so it's recommended to use the methods
	// defined on this interface instead.
	ResponseWriter() http.ResponseWriter
	Params() map[string]string
}

// BasicRequestContext is a basic implementation of RequestContext. It can be embedded in
// other types to provide a default implementation of the RequestContext interface.
type BasicReqContext struct {
	req    *http.Request
	w      http.ResponseWriter
	ctx    context.Context
	params map[string]string
}

var _ RequestContext = (*BasicReqContext)(nil)

func NewRequestContext(req *http.Request, w http.ResponseWriter, params map[string]string) *BasicReqContext {
	return &BasicReqContext{
		req:    req,
		w:      w,
		params: params,
	}
}

func (r *BasicReqContext) Request() *http.Request {
	return r.req
}
func (r *BasicReqContext) ResponseWriter() http.ResponseWriter {
	return r.w
}

func (r *BasicReqContext) Params() map[string]string {
	return r.params
}
