package fernet

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRouter(t *testing.T) {
	router := New[int]()

	handler := func(res Response, req *Request[int]) {
		res.WriteStatus(http.StatusCreated)
		res.Header().Set("Content-Type", "application/json")
		res.Write([]byte(`{"foo": "bar"}`))
	}

	tests := map[string]struct {
		routerFn func(string, Handler[int])
		method   string
	}{
		"GET":    {method: http.MethodGet, routerFn: router.Get},
		"POST":   {method: http.MethodPost, routerFn: router.Post},
		"PUT":    {method: http.MethodPut, routerFn: router.Put},
		"PATCH":  {method: http.MethodPatch, routerFn: router.Patch},
		"DELETE": {method: http.MethodDelete, routerFn: router.Delete},
	}

	for testName, tc := range tests {
		t.Run(testName, func(t *testing.T) {
			tc.routerFn("/foo", handler)

			res := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, "/foo", nil)

			router.ServeHTTP(res, req)

			require.Equal(t, http.StatusCreated, res.Code)
			require.Equal(t, "application/json", res.Header().Get("Content-Type"))
			require.Equal(t, `{"foo": "bar"}`, res.Body.String())
		})
	}
}

func TestRouter_Root(t *testing.T) {
	router := New[int]()

	router.Get("/", func(res Response, req *Request[int]) {
		res.WriteStatus(http.StatusCreated)
		res.Header().Set("Content-Type", "application/json")
		res.Write([]byte(`{"foo": "bar"}`))
	})

	require.Equal(t, 1, len(router.routes))
	require.Equal(t, "GET", router.routes[0].Method)
	require.Equal(t, "/", router.routes[0].Raw)

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusCreated, res.Code)
	require.Equal(t, "application/json", res.Header().Get("Content-Type"))
	require.Equal(t, `{"foo": "bar"}`, res.Body.String())
}

func TestRouter_Missing(t *testing.T) {
	router := New[int]()

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusNotFound, res.Code)
}

type contextKey struct{}
type beforeContextKey struct{}

func TestRouter_Metal(t *testing.T) {
	router := New[int]()
	router.UseMetal(func(w http.ResponseWriter, r *http.Request, h http.Handler) {
		ctx := context.WithValue(r.Context(), contextKey{}, "bar")
		h.ServeHTTP(w, r.WithContext(ctx))
	})
	router.UseMetal(func(w http.ResponseWriter, r *http.Request, h http.Handler) {
		require.Equal(t, "bar", r.Context().Value(contextKey{}))
		w.Header().Set("x-foo", "bar")
		h.ServeHTTP(w, r)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	router.ServeHTTP(res, req)

	require.Equal(t, "bar", res.Header().Get("x-foo"))
}

func TestRouter_Before(t *testing.T) {
	router := New[int]()
	router.UseMetal(func(w http.ResponseWriter, r *http.Request, h http.Handler) {
		ctx := context.WithValue(r.Context(), contextKey{}, "bar")
		h.ServeHTTP(w, r.WithContext(ctx))
	})
	router.Use(func(res Response, req *Request[int], next Handler[int]) {
		require.Equal(t, "bar", req.Context().Value(contextKey{}))
		ctx := context.WithValue(req.Context(), beforeContextKey{}, "baz")

		next(res, req.WithContext(ctx))
	})
	router.Use(func(res Response, req *Request[int], next Handler[int]) {
		require.Equal(t, "bar", req.Context().Value(contextKey{}))
		require.Equal(t, "baz", req.Context().Value(beforeContextKey{}))
		res.Header().Set("x-metal", "bar")
		res.Header().Set("x-before", "baz")

		next(res, req)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	router.Get("/", func(res Response, req *Request[int]) {
		res.Write([]byte("Hello world"))
	})

	router.ServeHTTP(res, req)

	require.Equal(t, "bar", res.Header().Get("x-metal"))
	require.Equal(t, "baz", res.Header().Get("x-before"))
}

func TestRouter_Params(t *testing.T) {
	router := New[int]()

	router.Get("/hello/:name", func(res Response, req *Request[int]) {
		res.Write([]byte(
			fmt.Sprintf("Hello %s", req.Param("name")),
		))
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/hello/fox", nil)

	router.ServeHTTP(res, req)
	require.Equal(t, "Hello fox", res.Body.String())
}
