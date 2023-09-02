package fernet

import (
	"errors"
	"net/http"
)

type ResponseWriter interface {
	// Status returns the status to be written to the client
	Status() int
	Flush() (int, error)
	Clear()
	http.ResponseWriter
}

var ErrAlreadyFlushed error = errors.New("response has already been flushed")

// responseWriter implements the http.responseWriter interface and exposes
// additional information about the response like the status code and number of
// bytes written.
type responseWriter struct {
	status  int
	body    []byte
	rw      http.ResponseWriter
	flushed bool
}

var _ http.ResponseWriter = (*responseWriter)(nil)

func newResponseWriter(rw http.ResponseWriter) *responseWriter {
	return &responseWriter{
		status: 200,
		body:   []byte{},
		rw:     rw,
	}
}

// WriteHeader writes the status code of the response.
func (r *responseWriter) WriteHeader(status int) {
	r.status = status
}

// Write implements the http.ResponseWriter interface and buffers the bytes to
// be written.
func (r *responseWriter) Write(b []byte) (int, error) {
	r.body = append(r.body, b...)

	return len(b), nil
}

// Header represents the header map of the response.
func (r *responseWriter) Header() http.Header {
	return r.rw.Header()
}

// Status returns the status code of the response.
func (r *responseWriter) Status() int {
	return r.status
}

// Flush writes the buffered bytes to the underlying http.ResponseWriter.
func (r *responseWriter) Flush() (int, error) {
	if r.flushed {
		return 0, ErrAlreadyFlushed
	}

	r.flushed = true
	r.rw.WriteHeader(r.status)
	return r.rw.Write(r.body)
}

// Clear resets the body that would be written to the client
func (r *responseWriter) Clear() {
	r.body = []byte{}
}
