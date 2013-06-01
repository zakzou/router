package router

import (
	"net/http"
)

const (
	HookBefore = "hook.before"
	HookAfter  = "hook.after"
)

// middleware
type Middleware interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type MiddlewareFunc func(http.ResponseWriter, *http.Request)

func (m MiddlewareFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m(w, r)
}

// hook
type Hook interface {
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type HookFunc func(http.ResponseWriter, *http.Request)

func (h HookFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h(w, r)
}
