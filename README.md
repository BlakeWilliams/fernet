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
  app.Use(func(w fernet.ResponseWriter, r *fernet.Request[AppData], next fernet.Handler) {
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
