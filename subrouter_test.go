package fernet

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

type PostData struct {
	ID int
}

func (p *PostData) FromRequest(ctx context.Context, r *RootRequestContext) bool {
	p.ID = 1
	return true
}

type CommentData struct {
	ID int
}

func (c *CommentData) FromRequest(ctx context.Context, r *RootRequestContext) bool {
	stringID := r.Params()["id"]
	id, err := strconv.Atoi(stringID)
	if err != nil {
		r.Response().WriteHeader(http.StatusBadRequest)
		return false
	}

	c.ID = id
	return true
}

func TestSubRouter(t *testing.T) {
	router := New(WithBasicRequestContext)
	subrouter := NewSubRouter(router, &PostData{})

	handler := func(ctx context.Context, r *RootRequestContext, postData *PostData) {
		r.Response().Header().Set("Content-Type", "application/json")
		r.Response().WriteHeader(http.StatusCreated)
		_, _ = r.Response().Write([]byte(`{"foo": "bar"}`))
	}

	tests := map[string]struct {
		routerFn func(string, SubRouterHandler[*RootRequestContext, *PostData])
		method   string
	}{
		"GET":    {method: http.MethodGet, routerFn: subrouter.Get},
		"POST":   {method: http.MethodPost, routerFn: subrouter.Post},
		"PUT":    {method: http.MethodPut, routerFn: subrouter.Put},
		"PATCH":  {method: http.MethodPatch, routerFn: subrouter.Patch},
		"DELETE": {method: http.MethodDelete, routerFn: subrouter.Delete},
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

func TestSubRouter_Routing(t *testing.T) {
	router := New(WithBasicRequestContext)

	subrouter := NewSubRouter(router, &PostData{})
	subrouter.Match("GET", "/", func(ctx context.Context, r *RootRequestContext, p *PostData) {
		r.Response().Header().Set("Content-Type", "application/json")
		r.Response().WriteHeader(http.StatusCreated)
		_, _ = r.Response().Write([]byte(fmt.Sprintf(`{"id": "%d"}`, p.ID)))
	})

	require.Equal(t, 1, len(router.routes))
	require.Equal(t, "GET", router.routes[0].Method)
	require.Equal(t, "/", router.routes[0].Path)

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusCreated, res.Code)
	require.Equal(t, "application/json", res.Header().Get("Content-Type"))
	require.Equal(t, `{"id": "1"}`, res.Body.String())
}

func Test_SubRouterSubRouter(t *testing.T) {
	router := New(WithBasicRequestContext)

	subrouter := NewSubRouter(router, &PostData{})
	subsubrouter := NewSubRouter(subrouter, &CommentData{})

	subsubrouter.Match("GET", "/comments/:id", func(ctx context.Context, r *RootRequestContext, c *CommentData) {
		r.Response().Header().Set("Content-Type", "application/json")
		r.Response().WriteHeader(http.StatusCreated)
		_, _ = r.Response().Write([]byte(fmt.Sprintf(`{"id": "%d"}`, c.ID)))
	})

	require.Equal(t, 1, len(router.routes))
	require.Equal(t, "GET", router.routes[0].Method)
	require.Equal(t, "/comments/:id", router.routes[0].Path)

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/comments/4", nil)

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusCreated, res.Code)
	require.Equal(t, "application/json", res.Header().Get("Content-Type"))
	require.Equal(t, `{"id": "4"}`, res.Body.String())
}

func Test_FromRequestFalse(t *testing.T) {
	router := New(WithBasicRequestContext)

	subrouter := NewSubRouter(router, &CommentData{})
	subrouter.Match("GET", "/comments/:id", func(ctx context.Context, r *RootRequestContext, p *CommentData) {
		r.Response().Header().Set("Content-Type", "application/json")
		r.Response().WriteHeader(http.StatusCreated)
		_, _ = r.Response().Write([]byte(fmt.Sprintf(`{"id": "%d"}`, p.ID)))
	})

	require.Equal(t, 1, len(router.routes))
	require.Equal(t, "GET", router.routes[0].Method)
	require.Equal(t, "/comments/:id", router.routes[0].Path)

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/comments/wow", nil)

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusBadRequest, res.Code)
}

func Test_SubRouterGroupPrefix(t *testing.T) {
	router := New(WithBasicRequestContext)

	subrouter := NewSubRouter(router, &PostData{})
	group := subrouter.Group("/comments")
	group.RawMatch(http.MethodGet, "/testing", func(ctx context.Context, r *RootRequestContext) {})
	subgroup := group.Group("/sub")
	fmt.Println("HERE")
	subgroup.Get("/get", func(ctx context.Context, r *RootRequestContext, p *PostData) {})

	res := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/comments/testing", nil)
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)

	res = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/comments/sub/get", nil)
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)
}
