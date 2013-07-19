router
======

A HTTP router implemented in Golang


## Documentation

### initialize
```go
r := router.NewRouter()
```

### register middleware for all route
```go
r.MiddlewareFunc(func(w http.ResponseWriter, r *http.Request) {
    // do something
})
```

### register middleware for a route
```go
r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
    // do something
}).MiddlewareFunc(func(w http.ResponseWriter, r *http.Request) {
    // do something
}).Methods("GET", "POST")
```

### register hook
```go
r.HookFunc(route.HookBeforeRouter, func(w http.ResponseWriter, r *http.Request) {
    // do something
})

r.HookFunc(route.HookAfterDispatch, func(w http.ResponseWriter, r *http.Request) {
    // do something
})
```

### URL For
`UrlFor()` method lest you dynamically create URLs for a named route
```go
r.HandleFunc("/user/profile/query/<int:user_id>/", func(w http.ResponseWriter, r *http.Request) {
    // do something
}).Name("profile")

if urls, ok := r.UrlFor("profile", map[string]interface{}{"user_id": 100001}); ok {
    println(urls)
}
```


### run
```go
if err := http.ListenAndServe(":9090", r); err != nil {
    // do something
}
```


License
==============

MIT: http://rem.mit-license.org
