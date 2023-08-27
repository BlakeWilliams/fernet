package fernet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroup(t *testing.T) {
	router := New[int]()
	group := router.Group("/api/")

	handler := func(res Response, req *Request[int]) {
		res.WriteStatus(http.StatusCreated)
		res.Header().Set("Content-Type", "application/json")
		res.Write([]byte(`{"foo": "bar"}`))
	}

	tests := map[string]struct {
		routerFn func(string, Handler[int])
		method   string
	}{
		"GET":    {method: http.MethodGet, routerFn: group.Get},
		"POST":   {method: http.MethodPost, routerFn: group.Post},
		"PUT":    {method: http.MethodPut, routerFn: group.Put},
		"PATCH":  {method: http.MethodPatch, routerFn: group.Patch},
		"DELETE": {method: http.MethodDelete, routerFn: group.Delete},
	}

	for testName, tc := range tests {
		t.Run(testName, func(t *testing.T) {
			tc.routerFn("/foo", handler)

			res := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, "/api/foo", nil)

			router.ServeHTTP(res, req)

			require.Equal(t, http.StatusCreated, res.Code)
			require.Equal(t, "application/json", res.Header().Get("Content-Type"))
			require.Equal(t, `{"foo": "bar"}`, res.Body.String())
		})
	}
}

func TestGroup_Middleware(t *testing.T) {
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

	group := router.Group("/api")

	group.Use(func(res Response, req *Request[int], next Handler[int]) {
		require.Equal(t, "bar", req.Context().Value(contextKey{}))
		require.Equal(t, "baz", req.Context().Value(beforeContextKey{}))
		res.Header().Set("x-group", "yolo")

		next(res, req)
	})

	group.Get("/foo", func(res Response, req *Request[int]) {
		res.Write([]byte("Hello world"))
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/foo", nil)
	router.ServeHTTP(res, req)

	require.Equal(t, "bar", res.Header().Get("x-metal"))
	require.Equal(t, "baz", res.Header().Get("x-before"))
	require.Equal(t, "yolo", res.Header().Get("x-group"))
}

func TestGroup_NestedGroup(t *testing.T) {
	router := New[int]()
	group := router.Group("/api")
	subgroup := group.Group("/v1")

	group.Use(func(res Response, req *Request[int], next Handler[int]) {
		ctx := context.WithValue(req.Context(), contextKey{}, "foo")
		next(res, req.WithContext(ctx))
	})

	subgroup.Use(func(res Response, req *Request[int], next Handler[int]) {
		require.Equal(t, "foo", req.Context().Value(contextKey{}))
		res.Header().Set("x-subgroup", "v1")
		next(res, req)
	})

	subgroup.Get("/foo", func(res Response, req *Request[int]) {
		res.Write([]byte("Hello world"))
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/foo", nil)
	router.ServeHTTP(res, req)

	require.Equal(t, "v1", res.Header().Get("x-subgroup"))
}
