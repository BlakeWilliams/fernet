package fernet

import "net/http"

// Response represents the response to a request. It is used to set the status
// code and write the response body. It does not write the response immediately,
// by default, but instead buffers the response until the request is finished.
type Response interface {
	// Returns the set status code.
	Status(int)
	WriteStatus(int)
	Header() http.Header
	Write(b []byte)
}

// response is the default implementation of the Response interface.
type response struct {
	status int
	header http.Header
	body   []byte
}

var _ Response = (*response)(nil)

func (r *response) Status(status int) {
	r.status = status
}
func (r *response) WriteStatus(status int) {
	r.status = status
}
func (r *response) Header() http.Header {
	return r.header
}
func (r *response) Write(b []byte) {
	r.body = append(r.body, b...)
}
