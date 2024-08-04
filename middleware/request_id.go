package middleware

import (
	"context"

	"github.com/blakewilliams/fernet"
	"github.com/google/uuid"
)

type requestIDKey struct{}

// RequestID is a middleware that retrieves the request ID from the
// "X-Request-ID" header. If the header is not set, a new UUID is generated and
// set as the request ID.
//
// The request ID is set on the context and the "X-Request-ID" header. The
// RequestIDFromContext function can be used to retrieve the request ID.
func RequestID[RC fernet.RequestContext]() func(context.Context, RC, fernet.Handler[RC]) {
	return func(ctx context.Context, rc RC, next fernet.Handler[RC]) {
		requestID := rc.Request().Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
		}

		ctx = context.WithValue(ctx, requestIDKey{}, requestID)
		rc.Response().Header().Set("X-Request-ID", requestID)

		next(ctx, rc)
	}
}

// RequestIDFromContext returns the request ID from the context if it exists.
func RequestIDFromContext(ctx context.Context) (string, bool) {
	if requestID, ok := ctx.Value(requestIDKey{}).(string); ok {
		return requestID, true
	}

	return "", false
}
