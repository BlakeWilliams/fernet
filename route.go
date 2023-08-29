package fernet

import (
	"net/http"
	"strings"
)

type route[T ReqRes] struct {
	Method  string
	Raw     string
	parts   []string
	handler Handler[T]
}

func (r *route[C]) Match(req *http.Request) (bool, map[string]string) {
	if r.Method != req.Method {
		return false, nil
	}

	reqParts := normalizeRoutePath(req.URL.Path)

	if len(r.parts) != len(reqParts) {
		return false, nil
	}

	params := make(map[string]string)

	for i, part := range r.parts {
		if strings.HasPrefix(part, ":") {
			params[part[1:]] = reqParts[i]
		} else if part != reqParts[i] {
			return false, nil
		}
	}

	return true, params
}

func newRoute[T ReqRes](method string, path string, handler Handler[T]) *route[T] {
	parts := normalizeRoutePath(path)

	// TODO better support for `/`, remove double `//`
	return &route[T]{
		Method:  method,
		Raw:     path,
		parts:   parts,
		handler: handler,
	}
}

func normalizeRoutePath(path string) []string {
	path = strings.TrimPrefix(path, "/")
	return strings.Split(path, "/")
}
