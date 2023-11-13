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

    // Use is used to add fernet based middleware to the application.
    app.Use(func(ctx context.Context, r *RequestContext, next fernet.Handler[RequestContext]) {
        // Do something before the request is handled.
        next(ctx, r)
        // Do something after the request is handled.
    })

    // Fernet routing uses : to define named parameters in the path. Wildcards are also supported via *.
    app.Get("/hello/:name", func(ctx context.Context, r *RequestContext) {
        r.WriteString(http.StatusOK, fmt.Sprintf("Hello %s", rc.Params()["name"]))
    })

    // Handle 404s by defining a catch-all route.
    app.Get("*", func(ctx context.Context, r *RequestContext) {
        r.WriteString(http.StatusNotFound, "Not Found")
    }

    app.ListenAndServe(":3200")
}
```

## SubRouters

SubRouters are similar to the regular router and `Group` types, but SubRouter handlers accept a third argument that implements `FromRequest`. This allows you to define a type that can be used to extract data from the request and pass it to the handler. e.g.

```go
// Define a type that implements FromRequest and can store the team record.
type TeamData struct { Team *Team }

// Implement the FromRequest method. If it returns false, the handler will not
// be called. If it returns true, the request will be processed as normal.
func (td *TeamData) FromRequest(ctx context.Context, rc *AppRequestContext) error {
    td.Team = rc.TeamRepository.Find(ctx, rc.Params["team_id"])
    // Handle missing data
    if td.Team == nil {
        rc.Render404()
        return false
    }

    // Handle authorization
    if rc.TeamRepository.IsMember(ctx, rc.CurrentUser, td.Team) {
        rc.Render403()
        return false
    }

    return true
}

// Define a handler that accepts the TeamData type.
func Show(ctx context.Context, rc *AppRequestContext, td *TeamData) {
    rc.RenderJSON(http.StatusOK, td.Team)
}

// Setup the router and subrouter.
router := fernet.New(func(r *fernet.RequestContext) *AppRequestContext {
    return &AppRequestContext{RequestContext: r}
})

teamsRouter := app.SubRouter(router, *TeamData{})
teamsRouter.Get("/teams/:team_id", Show)
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

## Metal

Fernet provides `Metal` handlers, which operate against net/http `Request` and
`ResponseWriter` types. These handlers are useful for integrating with existing
middleware and libraries or modifying the request/response before it is passed
to the `RequestContext` handler.

- `metal.MethodRewrite` - Rewrites the HTTP method based on the value of the `_method` form value.
