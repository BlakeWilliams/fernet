package fernet

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGroup(t *testing.T) {
	router := New(WithBasicRequestContext)
	group := router.Group("/api/")

	handler := func(ctx context.Context, r *RootRequestContext) {
		r.Response().Header().Set("Content-Type", "application/json")
		r.Response().WriteHeader(http.StatusCreated)
		_, _ = r.Response().Write([]byte(`{"foo": "bar"}`))
	}

	tests := map[string]struct {
		routerFn func(string, Handler[*RootRequestContext])
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
	router := New(WithBasicRequestContext)
	router.Use(func(ctx context.Context, r *RootRequestContext, next Handler[*RootRequestContext]) {
		ctx = context.WithValue(ctx, beforeContextKey{}, "baz")

		next(ctx, r)
	})
	router.Use(func(ctx context.Context, r *RootRequestContext, next Handler[*RootRequestContext]) {
		require.Equal(t, "baz", ctx.Value(beforeContextKey{}))
		r.Response().Header().Set("x-before", "baz")

		next(ctx, r)
	})

	group := router.Group("/api")

	group.Use(func(ctx context.Context, r *RootRequestContext, next Handler[*RootRequestContext]) {
		require.Equal(t, "baz", ctx.Value(beforeContextKey{}))
		r.Response().Header().Set("x-group", "yolo")

		next(ctx, r)
	})

	group.Get("/foo", func(ctx context.Context, r *RootRequestContext) {
		_, _ = r.Response().Write([]byte("Hello world"))
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/foo", nil)
	router.ServeHTTP(res, req)

	require.Equal(t, "baz", res.Header().Get("x-before"))
	require.Equal(t, "yolo", res.Header().Get("x-group"))
}

func TestGroup_NestedGroup(t *testing.T) {
	router := New(WithBasicRequestContext)
	group := router.Group("/api")
	subgroup := group.Group("/v1")

	group.Use(func(ctx context.Context, r *RootRequestContext, next Handler[*RootRequestContext]) {
		ctx = context.WithValue(ctx, contextKey{}, "foo")
		next(ctx, r)
	})

	subgroup.Use(func(ctx context.Context, r *RootRequestContext, next Handler[*RootRequestContext]) {
		require.Equal(t, "foo", ctx.Value(contextKey{}))
		r.Response().Header().Set("x-subgroup", "v1")
		next(ctx, r)
	})

	subgroup.Get("/foo", func(ctx context.Context, r *RootRequestContext) {
		_, _ = r.Response().Write([]byte("Hello world"))
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/foo", nil)
	router.ServeHTTP(res, req)

	require.Equal(t, "v1", res.Header().Get("x-subgroup"))
}

func TestGroup_PrefixRoot(t *testing.T) {
	router := New(WithBasicRequestContext)

	group := router.Group("/foo")
	group.Get("/", func(ctx context.Context, r *RootRequestContext) {
		r.Response().WriteHeader(http.StatusOK)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/foo", nil)
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
}
