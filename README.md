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
r.HookFunc(route.HookBefore, func(w http.ResponseWriter, r *http.Request) {
    // do something
})

r.HookFunc(route.HookAfter, func(w http.ResponseWriter, r *http.Request) {
    // do something
})
```


### run
```go
if err := http.ListenAndServe(":9090", r); err != nil {
    // do something
}
```
