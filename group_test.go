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

	handler := func(ctx context.Context, r *BasicReqContext) {
		r.ResponseWriter().Header().Set("Content-Type", "application/json")
		r.ResponseWriter().WriteHeader(http.StatusCreated)
		_, _ = r.ResponseWriter().Write([]byte(`{"foo": "bar"}`))
	}

	tests := map[string]struct {
		routerFn func(string, Handler[*BasicReqContext])
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
	router.UseMetal(func(w http.ResponseWriter, r *http.Request, h http.Handler) {
		ctx := context.WithValue(r.Context(), contextKey{}, "bar")
		h.ServeHTTP(w, r.WithContext(ctx))
	})
	router.Use(func(ctx context.Context, r *BasicReqContext, next Handler[*BasicReqContext]) {
		require.Equal(t, "bar", ctx.Value(contextKey{}))
		ctx = context.WithValue(ctx, beforeContextKey{}, "baz")

		next(ctx, r)
	})
	router.Use(func(ctx context.Context, r *BasicReqContext, next Handler[*BasicReqContext]) {
		require.Equal(t, "bar", ctx.Value(contextKey{}))
		require.Equal(t, "baz", ctx.Value(beforeContextKey{}))
		r.ResponseWriter().Header().Set("x-metal", "bar")
		r.ResponseWriter().Header().Set("x-before", "baz")

		next(ctx, r)
	})

	group := router.Group("/api")

	group.Use(func(ctx context.Context, r *BasicReqContext, next Handler[*BasicReqContext]) {
		require.Equal(t, "bar", ctx.Value(contextKey{}))
		require.Equal(t, "baz", ctx.Value(beforeContextKey{}))
		r.ResponseWriter().Header().Set("x-group", "yolo")

		next(ctx, r)
	})

	group.Get("/foo", func(ctx context.Context, r *BasicReqContext) {
		_, _ = r.ResponseWriter().Write([]byte("Hello world"))
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/foo", nil)
	router.ServeHTTP(res, req)

	require.Equal(t, "bar", res.Header().Get("x-metal"))
	require.Equal(t, "baz", res.Header().Get("x-before"))
	require.Equal(t, "yolo", res.Header().Get("x-group"))
}

func TestGroup_NestedGroup(t *testing.T) {
	router := New(WithBasicRequestContext)
	group := router.Group("/api")
	subgroup := group.Group("/v1")

	group.Use(func(ctx context.Context, r *BasicReqContext, next Handler[*BasicReqContext]) {
		ctx = context.WithValue(ctx, contextKey{}, "foo")
		next(ctx, r)
	})

	subgroup.Use(func(ctx context.Context, r *BasicReqContext, next Handler[*BasicReqContext]) {
		require.Equal(t, "foo", ctx.Value(contextKey{}))
		r.ResponseWriter().Header().Set("x-subgroup", "v1")
		next(ctx, r)
	})

	subgroup.Get("/foo", func(ctx context.Context, r *BasicReqContext) {
		_, _ = r.ResponseWriter().Write([]byte("Hello world"))
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/foo", nil)
	router.ServeHTTP(res, req)

	require.Equal(t, "v1", res.Header().Get("x-subgroup"))
}
