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

// AppData is used to hold global request data used by middleware, handlers, and
// templates.
type AppData struct {
  currentUser *User
}

func main() {
  app := fernet.New[AppData]()

  // UseMetal is used to add Go http based middleware to the application.
  app.UseMetal(func(w http.ResponseWriter, r *http.Request, next http.Handler) {
    // Do something before the request is handled.
    next.ServeHTTP(w, r)
    // Do something after the request is handled.
  })

// Use is used to add fernet based middleware to the application.
  app.Use(func(w fernet.ResponseWriter, r *fernet.Request[AppData], next fernet.Handler[AppData]) {
    // Do something before the request is handled.
    next(w, r)
    // Do something after the request is handled.
  })

  app.Get("/", func(w fernet.ResponseWriter, r *fernet.Request[Appdata]) {
    fmt.Fprint(w, "Hello, World!")
  })

  app.ListenAndServe(":3200")
}
```

## Groups

Groups are used to group routes together and apply middleware common only to those groups and subgroups

```go
type AppData struct {
  currentUser *User
}

router := fernet.New[AppData]()

authGroup := router.Group("")
authGroup.Use(func(w fernet.ResponseWriter, r *fernet.Request[AppData], next fernet.Handler[AppData]) {
  if r.AppData.currentUser == nil {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }

  next(w, r)
})

adminGroup := authGroup.Group("/admin")
adminGroup.Use(func(w fernet.ResponseWriter, r *fernet.Request[AppData], next fernet.Handler[AppData]) {
  if r.AppData.currentUser == nil || r.AppData.currentUser.Role != "admin" {
    w.WriteHeader(http.StatusUnauthorized)
    return
  }

  next(w, r)
})
```

## `Registrable` and Controllers

Controllers are an often used pattern to group related routes together, typically with shared middleware and data requirements. Fernet provides a `Registrable` interface to make it easy to register controllers with your application.

```go
type AppData struct {
  currentUser *User
}

type UsersController struct {
  db *sql.DB
}

// Register is used to register the controller routes with the application. This
// simple abstraction makes it easy to extend fernet and your routing layer.
//
// For example, you could create your own router type that wraps `app` and adds
// behavior or encapsulates data.
func (c *UsersController) Register(app *fernet.App[AppData]) {
  app.Get("/users", c.Index)
  app.Get("/users/:id", c.Show)
}

func (c *UsersController) Index(w fernet.ResponseWriter, r *fernet.Request[AppData]) {
  // ...
}

func (c *UsersController) Show(w fernet.ResponseWriter, r *fernet.Request[AppData]) {
  // ...
}

func main() {
  app := fernet.New[AppData]()

  db, err := sql.Open("postgres", "...")
  if err != nil {
    panic(err)
  }

  app.Register(&UsersController{db: db})

  app.ListenAndServe(":3200")
}
```
