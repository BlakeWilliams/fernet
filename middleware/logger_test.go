package middleware

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/blakewilliams/fernet"
	"github.com/stretchr/testify/require"
)

func TestLogger(t *testing.T) {
	var b bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&b, nil))
	router := fernet.New(func(r fernet.RequestContext) fernet.RequestContext {
		return r
	})
	router.Use(Logger[fernet.RequestContext](logger))
	router.Get("/:name", func(ctx context.Context, r fernet.RequestContext) {
		r.Response().WriteHeader(http.StatusAccepted)
	})

	req := httptest.NewRequest(http.MethodGet, "/fox", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	lines := strings.Split(b.String(), "\n")
	require.Len(t, lines, 3) // Two log lines, empty newline

	reqLine := lines[0]
	require.Contains(t, reqLine, `"method":"GET"`)
	require.Contains(t, reqLine, `"path":"/fox"`)
	require.Contains(t, reqLine, `"route":"/:name"`)

	resLine := lines[1]
	require.Contains(t, resLine, `"status":202`)
	require.Contains(t, resLine, `"method":"GET"`)
	require.Contains(t, resLine, `"path":"/fox"`)
	require.Contains(t, resLine, `"route":"/:name"`)
	require.NotContains(t, resLine, `"request_id":`)
}

func TestLogger_RequestID(t *testing.T) {
	var b bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&b, nil))
	router := fernet.New(func(r fernet.RequestContext) fernet.RequestContext {
		return r
	})
	router.Use(
		RequestID[fernet.RequestContext](),
		Logger[fernet.RequestContext](logger),
	)
	router.Get("/:name", func(ctx context.Context, r fernet.RequestContext) {
		r.Response().WriteHeader(http.StatusAccepted)
	})

	req := httptest.NewRequest(http.MethodGet, "/fox", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	lines := strings.Split(b.String(), "\n")
	require.Len(t, lines, 3) // Two log lines, empty newline

	reqLine := lines[0]
	require.Contains(t, reqLine, `"method":"GET"`)
	require.Contains(t, reqLine, `"path":"/fox"`)
	require.Contains(t, reqLine, `"route":"/:name"`)
	require.Contains(t, reqLine, `"request_id":`)

	resLine := lines[1]
	require.Contains(t, resLine, `"status":202`)
	require.Contains(t, resLine, `"method":"GET"`)
	require.Contains(t, resLine, `"path":"/fox"`)
	require.Contains(t, resLine, `"request_id":`)
}
