package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blakewilliams/fernet"
	"github.com/stretchr/testify/require"
)

func TestRewrite(t *testing.T) {
	router := fernet.New(func(r fernet.RequestContext) fernet.RequestContext { return r })
	router.Use(MethodRewrite)
	router.Delete("/", func(ctx context.Context, rc fernet.RequestContext) {})

	req := httptest.NewRequest(http.MethodDelete, "/", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Result().StatusCode)
}

func Test_RewritePost(t *testing.T) {
	router := fernet.New(func(r fernet.RequestContext) fernet.RequestContext { return r })
	router.Use(MethodRewrite[fernet.RequestContext])
	router.Delete("/", func(ctx context.Context, rc fernet.RequestContext) {})

	req := httptest.NewRequest(http.MethodPost, "/", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusNotFound, res.Result().StatusCode)
}
