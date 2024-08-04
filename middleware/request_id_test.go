package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/blakewilliams/fernet"
	"github.com/stretchr/testify/require"
)

func TestRequestIDMiddleware(t *testing.T) {
	router := fernet.New(func(r fernet.RequestContext) fernet.RequestContext { return r })
	router.Use(
		RequestID[fernet.RequestContext](),
	)
	router.Get("/:name", func(ctx context.Context, r fernet.RequestContext) {
		requestID, ok := RequestIDFromContext(ctx)
		require.True(t, ok)

		r.Response().WriteHeader(http.StatusAccepted)
		_, _ = r.Response().Write([]byte(requestID))
	})

	req := httptest.NewRequest(http.MethodGet, "/fox", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusAccepted, res.Code)
	require.NotEmpty(t, res.Body.String())
	require.Equal(t, res.Body.String(), res.Header().Get("X-Request-ID"))
}

func TestRequestIDMiddleware_ExistingRequestID(t *testing.T) {
	router := fernet.New(func(r fernet.RequestContext) fernet.RequestContext { return r })
	router.Use(
		RequestID[fernet.RequestContext](),
	)
	router.Get("/:name", func(ctx context.Context, r fernet.RequestContext) {
		requestID, ok := RequestIDFromContext(ctx)
		require.True(t, ok)

		r.Response().WriteHeader(http.StatusAccepted)
		_, _ = r.Response().Write([]byte(requestID))
	})

	req := httptest.NewRequest(http.MethodGet, "/fox", nil)
	req.Header.Set("X-Request-ID", "existing-request-id")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusAccepted, res.Code)
	require.Equal(t, "existing-request-id", res.Body.String())
	require.Equal(t, res.Body.String(), res.Header().Get("X-Request-ID"))
}
