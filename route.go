package fernet

import (
	"net/http"
	"strings"
)

type route[T RequestContext] struct {
	Method  string
	Path    string
	parts   []string
	handler Handler[T]
}

func (r *route[C]) match(req *http.Request) (bool, map[string]string) {
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

func newRoute[T RequestContext](method string, path string, handler Handler[T]) *route[T] {
	parts := normalizeRoutePath(path)

	// TODO better support for `/`, remove double `//`
	return &route[T]{
		Method:  method,
		Path:    path,
		parts:   parts,
		handler: handler,
	}
}

func normalizeRoutePath(path string) []string {
	path = strings.TrimPrefix(path, "/")
	return strings.Split(path, "/")
}
