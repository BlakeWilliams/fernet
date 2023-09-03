# Fernet

A simple Go framework for building web applications. Fernet uses generics to
provide convenient and type-safe APIs for your Handlers and Middleware.

## Getting Started

```go
package main

import (
    "fmt"
    "net/http"

    "github.com/blakewilliams/fernet"
)

// RequestContext is used to store data that is shared between middleware and
// handlers. Methods can be defined on this type to provide application specific
// functionality like rendering.
type RequestContext struct {
    currentUser *User
    fernet.RequestContext
}

// Implement a basic render string helper function.
func (r *RequestContext) RenderString(code int, s string) {
    r.ResponseWriter().WriteHeader(code)
    r.ResponseWriter().Write([]byte(s))
}

func main() {
    app := fernet.New(func(r *fernet.RequestContext) *RequestContext {
        return &RequestContext{RequestContext: r}
    })

    // UseMetal is used to add Go http based middleware to the application.
    app.UseMetal(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
        // Do something before the request is handled.
        next.ServeHTTP(w, r)
        // Do something after the request is handled.
    })

    // Use is used to add fernet based middleware to the application.
    app.Use(func(ctx context.Context, r *RequestContext, next fernet.Handler[RequestContext]) {
        // Do something before the request is handled.
        next(ctx, r)
        // Do something after the request is handled.
    })

    app.Get("/", func(ctx context.Context, r *RequestContext) {
        r.WriteString(http.StatusOK, "Hello World!")
    })

    app.ListenAndServe(":3200")
}
```

## Groups

Groups are used to group routes together and apply middleware common only to those groups and subgroups

```go
type RequestContext struct {
    currentUser *User
    fernet.RequestContext
}

func (r *RequestContext) RenderString(code int, s string) {
    r.ResponseWriter().WriteHeader(code)
    r.ResponseWriter().Write([]byte(s))
}

app := fernet.New(func(r *fernet.RequestContext) *RequestContext {
    return &RequestContext{RequestContext: r}
})

authGroup := router.Group("")
authGroup.Use(func(ctx context.Context, r *RequestContext, next fernet.Handler[RequestContext]) {
    if r.AppData.currentUser == nil {
        r.RenderString(http.StatusUnauthorized, "Unauthorized")
        return
    }

    next(ctx, r)
})

adminGroup := authGroup.Group("/admin")
adminGroup.Use(func(ctx context.Context, r *RequestContext, next fernet.Handler[RequestContext]) {
    if r.AppData.currentUser == nil || r.AppData.currentUser.Role != "admin" {
        r.RenderString(http.StatusUnauthorized, "Unauthorized")
        return
    }

    next(ctx, r)
})
```

## Middleware

Fernet provides a few middleware functions out of the box. Import the
`github.com/blakewilliams/fernet/middleware` package to use them.

- `middleware.ErrorHandler` - rescues panics and calls a `Handler[T]` to handle
  the error.
- `middleware.Logger` - logs requests and responses using slog.
- `middleware.MethodRewrite` - Rewrites the HTTP method based on the value of the `_method` form value.
