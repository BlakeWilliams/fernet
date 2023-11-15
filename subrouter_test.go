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

func TestController(t *testing.T) {
	router := New(WithBasicRequestContext)
	controller := NewController(router, &PostData{})

	handler := func(ctx context.Context, r *RootRequestContext, postData *PostData) {
		r.Response().Header().Set("Content-Type", "application/json")
		r.Response().WriteHeader(http.StatusCreated)
		_, _ = r.Response().Write([]byte(`{"foo": "bar"}`))
	}

	tests := map[string]struct {
		routerFn func(string, ControllerHandler[*RootRequestContext, *PostData])
		method   string
	}{
		"GET":    {method: http.MethodGet, routerFn: controller.Get},
		"POST":   {method: http.MethodPost, routerFn: controller.Post},
		"PUT":    {method: http.MethodPut, routerFn: controller.Put},
		"PATCH":  {method: http.MethodPatch, routerFn: controller.Patch},
		"DELETE": {method: http.MethodDelete, routerFn: controller.Delete},
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

func TestController_Routing(t *testing.T) {
	router := New(WithBasicRequestContext)

	controller := NewController(router, &PostData{})
	controller.Match("GET", "/", func(ctx context.Context, r *RootRequestContext, p *PostData) {
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

func Test_NestedController(t *testing.T) {
	router := New(WithBasicRequestContext)

	postController := NewController(router, &PostData{})
	commentController := NewController(postController, &CommentData{})

	commentController.Match("GET", "/comments/:id", func(ctx context.Context, r *RootRequestContext, c *CommentData) {
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

	commentController := NewController(router, &CommentData{})
	commentController.Match("GET", "/comments/:id", func(ctx context.Context, r *RootRequestContext, p *CommentData) {
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

func Test_ControllerGroupPrefix(t *testing.T) {
	router := New(WithBasicRequestContext)

	controller := NewController(router, &PostData{})
	group := controller.Group("/comments")
	group.RawMatch(http.MethodGet, "/testing", func(ctx context.Context, r *RootRequestContext) {})
	subgroup := group.Group("/sub")
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

type TrackingRequestContext struct {
	Chain []string
	RequestContext
}

func (t *TrackingRequestContext) AddToChain(s string) {
	t.Chain = append(t.Chain, s)
}

type TrackingData struct{}

func (t *TrackingData) FromRequest(ctx context.Context, r *TrackingRequestContext) bool {
	r.AddToChain("FromRequest")
	return true
}

func Test_ControllerMiddleware(t *testing.T) {
	var tracking *TrackingRequestContext
	router := New(func(r RequestContext) *TrackingRequestContext {
		tracking = &TrackingRequestContext{RequestContext: r, Chain: []string{"new"}}
		return tracking
	})

	router.Use(func(ctx context.Context, r *TrackingRequestContext, next Handler[*TrackingRequestContext]) {
		r.AddToChain("router use")
		next(ctx, r)
	})

	group := router.Group("/comments")
	group.Use(func(ctx context.Context, r *TrackingRequestContext, next Handler[*TrackingRequestContext]) {
		r.AddToChain("group use")
		next(ctx, r)
	})
	controller := NewController(group, &TrackingData{})
	controller.Use(func(ctx context.Context, r *TrackingRequestContext, next Handler[*TrackingRequestContext]) {
		r.AddToChain("controller use")
		next(ctx, r)
	})
	subGroup := controller.Group("/sub")
	subGroup.Use(func(ctx context.Context, r *TrackingRequestContext, next Handler[*TrackingRequestContext]) {
		r.AddToChain("subgroup use")
		next(ctx, r)
	})
	subGroup.Get("/best", func(ctx context.Context, r *TrackingRequestContext, p *TrackingData) {
		r.AddToChain("handler")
	})

	req := httptest.NewRequest("GET", "/comments/sub/best", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)

	require.Equal(
		t,
		[]string{"new", "router use", "group use", "controller use", "subgroup use", "FromRequest", "handler"},
		tracking.Chain,
		"expected the middleware, FromRequest, and handlers to be called in order",
	)
}
