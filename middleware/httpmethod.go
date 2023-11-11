package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/blakewilliams/fernet"
)

// MethodRewrite rewrites the HTTP method based on the _method parameter
// passed when the request type is POST. This is useful when working with HTTP
// forms since form only supports GET and POST methods.
func MethodRewrite[T fernet.RequestContext](ctx context.Context, rc T, next fernet.Next[T]) {
	if rc.Request().Method == http.MethodPost {
		if method := rc.Request().FormValue("_method"); method != "" {
			rc.Request().Method = strings.ToUpper(method)
		}
	}

	next(ctx, rc)
}
