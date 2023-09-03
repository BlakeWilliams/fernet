package middleware

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blakewilliams/fernet"
	"github.com/stretchr/testify/require"
)

func TestErrors(t *testing.T) {
	router := fernet.New(func(r fernet.RequestContext) fernet.RequestContext {
		return r
	})

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	router.Use(ErrorHandler(logger, func(ctx context.Context, r fernet.RequestContext, err any) {
		r.Response().WriteHeader(http.StatusInternalServerError)
		_, _ = r.Response().Write([]byte("something went wrong"))
	}))

	router.Get("/ok", func(ctx context.Context, r fernet.RequestContext) {
		_, _ = r.Response().Write([]byte("all good!"))
	})

	router.Get("/not-ok", func(ctx context.Context, r fernet.RequestContext) {
		panic("omg")
	})

	req := httptest.NewRequest(http.MethodGet, "/not-ok", nil)
	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusInternalServerError, res.Code)
	require.Equal(t, "something went wrong", res.Body.String())

	req = httptest.NewRequest(http.MethodGet, "/ok", nil)
	res = httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)
	require.Equal(t, "all good!", res.Body.String())
}
