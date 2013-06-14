package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var handlerOk = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world!")
}

var HandlerErr = func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusBadRequest)
}

func TestRouteOk(t *testing.T) {

	req, _ := http.NewRequest("GET", "/user/profile/10000/?test=value", nil)
	rw := httptest.NewRecorder()

	r := NewRouter()
	r.HandleFunc("/<string:one>/<string:two>/<int:three>/", handlerOk)
	r.ServeHTTP(rw, req)

	query := req.URL.Query()
	if one := query.Get("one"); one != "user" {
		t.Errorf("url param set to [%s]; want [%s]", one, "user")
	}
	if two := query.Get("two"); two != "profile" {
		t.Errorf("url param set to [%s]; want [%s]", two, "profile")
	}
	if three := query.Get("three"); three != "10000" {
		t.Errorf("url param set to [%s]; want [%s]", three, "10000")
	}
	if test := query.Get("test"); test != "value" {
		t.Errorf("url param set to [%s]; want [%s]", test, "value")
	}
}

func TestNotFound(t *testing.T) {
	req, _ := http.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	r := NewRouter()
	r.ServeHTTP(rw, req)

	if rw.Code != http.StatusNotFound {
		t.Errorf("Code set to [%v]; want [%v]", rw.Code, http.StatusNotFound)
	}
}

func TestRedirect(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test", nil)
	rw := httptest.NewRecorder()

	r := NewRouter()
	r.HandleFunc("/test/", handlerOk).StrictSlash(true)
	r.ServeHTTP(rw, req)

	if rw.Code != http.StatusMovedPermanently {
		t.Errorf("Code set to [%v]; want [%v]", rw.Code, http.StatusMovedPermanently)
	}
}

func TestMiddlreware(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test", nil)
	rw := httptest.NewRecorder()

	r := NewRouter()
	r.HandleFunc("/test/", handlerOk).MiddlewareFunc(HandlerErr)
	r.ServeHTTP(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Errorf("Code set to [%v]; want [%v]", rw.Code, http.StatusBadRequest)
	}
}

func TestHook(t *testing.T) {
	req, _ := http.NewRequest("GET", "/test", nil)
	rw := httptest.NewRecorder()

	r := NewRouter().HookFunc(HookBeforeRouter, HandlerErr)
	r.HandleFunc("/test/", handlerOk)
	r.ServeHTTP(rw, req)

	if rw.Code != http.StatusBadRequest {
		t.Errorf("Code set to [%v]; want [%v]", rw.Code, http.StatusBadRequest)
	}
}

func BenchmarkRouteHandler(b *testing.B) {
	r := NewRouter()
	r.HandleFunc("/", handlerOk)

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)
	}
}

func BenchmarkRouteHandlerParams(b *testing.B) {
	r := NewRouter()
	r.HandleFunc("/<string:name>/<int:user_id>/", handlerOk).StrictSlash(true)

	for i := 0; i < b.N; i++ {
		req, _ := http.NewRequest("GET", "/user/10000/", nil)
		rw := httptest.NewRecorder()
		r.ServeHTTP(rw, req)
	}
}

func BenchmarkServeMux(b *testing.B) {
	r := NewRouter()
	r.HandleFunc("/", handlerOk)

	req, _ := http.NewRequest("GET", "/", nil)
	rw := httptest.NewRecorder()

	for i := 0; i < b.N; i++ {
		r.ServeHTTP(rw, req)
	}
}
