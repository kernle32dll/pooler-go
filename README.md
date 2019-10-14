[![Build Status](https://travis-ci.com/kernle32dll/pooler-go.svg?branch=master)](https://travis-ci.com/kernle32dll/pooler-go)
[![GoDoc](https://godoc.org/github.com/kernle32dll/pooler-go?status.svg)](http://godoc.org/github.com/kernle32dll/pooler-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/kernle32dll/pooler-go)](https://goreportcard.com/report/github.com/kernle32dll/pooler-go)
[![codecov](https://codecov.io/gh/kernle32dll/pooler-go/branch/master/graph/badge.svg)](https://codecov.io/gh/kernle32dll/pooler-go)

# pooler-go

pooler-go is a small middleware, providing painless HTTP request-scoped object pooling capabilities.

What does that mean? pooler-go provides a simple middleware with an object pool. You take all the objects
you need, and pooler-go cleans them up after the request has ended.

Internally, pooler-go injects a holder into each request context. That key is user-definable - make sure
it is immutable, and don't mix pools for different objects.

Download:

```
go get github.com/kernle32dll/pooler-go
```

Detailed documentation can be found on [GoDoc](https://godoc.org/github.com/kernle32dll/pooler-go).

## Getting started

pooler-go is straight-forward to use. Initialize the middleware via `pooler.NewMiddleware(...)`,
use the middleware for your requests, and then retrieve new objects in your handler(s) via
`pooler.Get(...)`. No handing back required - pooler-go cleans up after the request has ended.

Take note to pass the **request** context (or a derived context) into the Get method.
Otherwise, pooler-go is unable to get the right internal holder.

```go
package main

import (
	"github.com/kernle32dll/pooler-go"

	"encoding/json"
	"log"
	"net/http"
	"time"
)

// User is a just sample struct for showcasing.
type User struct {
	Name string
	Time time.Time
}

const poolerKey = iota

func main() {
	router := http.NewServeMux()

	middleware := pooler.NewMiddleware(poolerKey, func() interface{} {
		log.Println("New object created")
		return &User{}
	})

	router.Handle("/user", middleware(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := pooler.Get(r.Context(), poolerKey).(*User)

			user.Name = "Björn Gerdau"
			user.Time = time.Now()

			decoder := json.NewEncoder(w)
			if err := decoder.Encode(user); err != nil {
				panic(err)
			}
		}),
	))

	log.Fatal(http.ListenAndServe(":8080", router))
}