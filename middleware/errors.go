package middleware

import (
	"context"
	"log/slog"

	"github.com/blakewilliams/fernet"
)

// ErrorHandler will catch panics in fernet applications and call the provided
// handler so that an error response can be rendered. It automatically calls
// `ResponseWriter.Clear` so partial responses aren't written to the client.
func ErrorHandler[T fernet.RequestContext](
	log *slog.Logger,
	handler func(ctx context.Context, rctx T, recovered any),
) func(context.Context, T, fernet.Handler[T]) {
	return func(ctx context.Context, rctx T, next fernet.Handler[T]) {
		defer func() {
			if rec := recover(); rec != nil {
				if err, ok := rec.(error); ok {
					log.Error("recovered in middleware", slog.String("error", err.Error()))
				} else {
					log.Error("recovered in middleware")
				}

				rctx.Response().Clear()
				handler(ctx, rctx, rec)
			}
		}()

		next(ctx, rctx)
	}
}
