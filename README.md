router
======

A HTTP router implemented in Golang


## Documentation

```go
package main

import (
	"fmt"
	"github.com/zakzou/router"
	"net/http"
)

type UserFilter struct {
	router.Middleware
}

func (f *UserFilter) BeforeRequest(w http.ResponseWriter, r *http.Request) {
	if user_id := r.URL.Query().Get("user_id"); user_id != "10000" {
		http.Redirect(w, r, "/", 301)
	}
}

func main() {
	rr := router.NewRouter()
	rr.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "hello world!")
	})

	ur := rr.SubRouter().Prefix("/user").Middlewares(new(UserFilter)).StrictSlash(true)
	ur.HandleFunc("/profile/query/<int:user_id>/", func(w http.ResponseWriter, r *http.Request) {
		user_id := r.URL.Query().Get("user_id")
		fmt.Fprintln(w, user_id, r.URL.Path)
	}).Methods("GET", "POST")

	http.ListenAndServe(":9090", rr)
}
```
