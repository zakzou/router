package router

import (
	"net/http"
)

type MiddlewareInterface interface {
	BeforeRequest(http.ResponseWriter, *http.Request)
	AfterRequest(http.ResponseWriter, *http.Request)
}

type Middleware struct {
}

func (m *Middleware) BeforeRequest(w http.ResponseWriter, r *http.Request) {
}

func (m *Middleware) AfterRequest(w http.ResponseWriter, r *http.Request) {
}
