package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/blakewilliams/fernet"
)

type responseStatusTracker struct {
	http.ResponseWriter
	status int
}

var _ http.ResponseWriter = (*responseStatusTracker)(nil)

func (r *responseStatusTracker) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseStatusTracker) Write(b []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}

	return r.ResponseWriter.Write(b)
}

func (r *responseStatusTracker) Header() http.Header {
	return r.ResponseWriter.Header()
}

func Logger[ReqCtx fernet.RequestContext](logger *slog.Logger) func(context.Context, ReqCtx, fernet.Handler[ReqCtx]) {
	return func(ctx context.Context, rctx ReqCtx, next fernet.Handler[ReqCtx]) {
		start := time.Now()

		logger.Info(
			"request started",
			slog.String("path", rctx.Request().URL.Path),
			slog.String("method", rctx.Request().Method),
			slog.String("route", rctx.MatchedPath()),
		)

		next(ctx, rctx)
		finished := time.Since(start)

		logger.Info(
			"request served",
			slog.String("path", rctx.Request().URL.Path),
			slog.String("method", rctx.Request().Method),
			slog.String("route", rctx.MatchedPath()),
			slog.Int("status", rctx.Response().Status()),
			slog.Int64("ms", finished.Milliseconds()),
		)
	}
}
