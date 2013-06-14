package router

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

var defaultHandlerFunc = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
	w.WriteHeader(http.StatusOK)
}

func TestDefaultMethod(t *testing.T) {
	route := NewRoute("/test/ok/", http.HandlerFunc(defaultHandlerFunc))
	if !route.supportsHttpMethod("GET") {
		t.Errorf("Should support GET")
	}
}

func TestOtherMethod(t *testing.T) {
	route := NewRoute("/test/ok/", http.HandlerFunc(defaultHandlerFunc)).Methods("POST")
	if !route.supportsHttpMethod("POST") {
		t.Errorf("Should support POST")
	}
}

func TestMatchPath(t *testing.T) {
	defaultHandler := http.HandlerFunc(defaultHandlerFunc)
	route := NewRoute("/test/ok/", defaultHandler).StrictSlash(true)
	if matched, ok := route.matches("/test/ok"); ok {
		if reflect.TypeOf(matched.handler) != reflect.TypeOf(http.RedirectHandler("/test/ok/", http.StatusMovedPermanently)) {
			t.Errorf("Should redirect")
		}
	} else {
		t.Errorf("Should match request")
	}
}

func TestMatchParams(t *testing.T) {
	route := NewRoute("/<string:value>/<int:user_id>/", http.HandlerFunc(defaultHandlerFunc))
	if matched, ok := route.matches("/user/10000/"); ok {
		params := matched.params
		if value := params.Get("value"); value != "user" {
			t.Errorf("url param set to [%s]; want [%s]", value, "user")
		}
		if userId := params.Get("user_id"); userId != "10000" {
			t.Errorf("url param set to [%s]; want [%s]", userId, "10000")
		}
	} else {
		t.Errorf("Should match request")
	}
}
