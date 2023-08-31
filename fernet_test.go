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
	router := New(WithBasicRequestContext)

	handler := func(ctx context.Context, r *RootRequestContext) {
		r.ResponseWriter().Header().Set("Content-Type", "application/json")
		r.ResponseWriter().WriteHeader(http.StatusCreated)
		_, _ = r.ResponseWriter().Write([]byte(`{"foo": "bar"}`))
	}

	tests := map[string]struct {
		routerFn func(string, Handler[*RootRequestContext])
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
	router := New(WithBasicRequestContext)

	router.Get("/", func(ctx context.Context, r *RootRequestContext) {
		r.ResponseWriter().Header().Set("Content-Type", "application/json")
		r.ResponseWriter().WriteHeader(http.StatusCreated)
		_, _ = r.ResponseWriter().Write([]byte(`{"foo": "bar"}`))
	})

	require.Equal(t, 1, len(router.routes))
	require.Equal(t, "GET", router.routes[0].Method)
	require.Equal(t, "/", router.routes[0].Path)

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusCreated, res.Code)
	require.Equal(t, "application/json", res.Header().Get("Content-Type"))
	require.Equal(t, `{"foo": "bar"}`, res.Body.String())
}

func TestRouter_Missing(t *testing.T) {
	router := New(WithBasicRequestContext)

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusNotFound, res.Code)
}

type contextKey struct{}
type beforeContextKey struct{}

func TestRouter_Metal(t *testing.T) {
	router := New(WithBasicRequestContext)
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
	router := New(WithBasicRequestContext)
	router.UseMetal(func(w http.ResponseWriter, r *http.Request, h http.Handler) {
		ctx := context.WithValue(r.Context(), contextKey{}, "bar")
		h.ServeHTTP(w, r.WithContext(ctx))
	})
	router.Use(func(ctx context.Context, r *RootRequestContext, next Handler[*RootRequestContext]) {
		require.Equal(t, "bar", ctx.Value(contextKey{}))
		ctx = context.WithValue(ctx, beforeContextKey{}, "baz")

		next(ctx, r)
	})
	router.Use(func(ctx context.Context, r *RootRequestContext, next Handler[*RootRequestContext]) {
		require.Equal(t, "bar", ctx.Value(contextKey{}))
		require.Equal(t, "baz", ctx.Value(beforeContextKey{}))
		r.ResponseWriter().Header().Set("x-metal", "bar")
		r.ResponseWriter().Header().Set("x-before", "baz")

		next(ctx, r)
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	router.Get("/", func(ctx context.Context, r *RootRequestContext) {
		_, _ = res.Write([]byte("Hello world"))
	})

	router.ServeHTTP(res, req)

	require.Equal(t, "bar", res.Header().Get("x-metal"))
	require.Equal(t, "baz", res.Header().Get("x-before"))
}

func TestRouter_Params(t *testing.T) {
	router := New(WithBasicRequestContext)

	router.Get("/hello/:name", func(ctx context.Context, r *RootRequestContext) {
		_, _ = r.ResponseWriter().Write([]byte(
			fmt.Sprintf("Hello %s", r.Params()["name"]),
		))
	})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/hello/fox", nil)

	router.ServeHTTP(res, req)
	require.Equal(t, "Hello fox", res.Body.String())
}

func WithBasicRequestContext(rctx RequestContext) *RootRequestContext {
	return rctx.(*RootRequestContext)
}
