package router

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

var HandlerOK = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "hello world")
	w.WriteHeader(http.StatusOK)
}

var HandlerErr = func(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "", http.StatusBadRequest)
}

type FilterUser struct {
	Middleware
}

func (m *FilterUser) AfterRequest(w http.ResponseWriter, r *http.Request) {
	if id := r.URL.Query().Get("user_id"); id == "10000" {
		http.Error(w, "", http.StatusUnauthorized)
	}
}

func TestRouteOK(t *testing.T) {
	r, _ := http.NewRequest("GET", "/user/profile/query/10000/?fields=id", nil)
	w := httptest.NewRecorder()

	rr := NewRouter()
	rr.HandleFunc("/user/profile/query/<int:user_id>/", HandlerOK)
	rr.ServeHTTP(w, r)

	user_id := r.URL.Query().Get("user_id")
	if user_id != "10000" {
		t.Errorf("url param set to [%s]; want [%s]", user_id, "10000")
	}
}

func TestNotFound(t *testing.T) {
	r, _ := http.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	rr := NewRouter()
	rr.ServeHTTP(w, r)
	if w.Code != http.StatusNotFound {
		t.Errorf("code set to [%v]; want [%v]", w.Code, http.StatusNotFound)
	}
}

func TestMiddleware(t *testing.T) {
	r, _ := http.NewRequest("GET", "/user/profile/query/10000/?fields=id", nil)
	w := httptest.NewRecorder()

	rr := NewRouter().Middlewares(&FilterUser{})
	rr.HandleFunc("/user/profile/query/<int:user_id>/", HandlerOK)
	rr.ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("code set to [%v]; want [%v]", w.Code, http.StatusNotFound)
	}
}

func BenchmarkRoute(b *testing.B) {
	rr := NewRouter()
	rr.HandleFunc("/", HandlerOK)

	for i := 0; i < b.N; i++ {
		r, _ := http.NewRequest("GET", "/", nil)
		w := httptest.NewRecorder()
		rr.ServeHTTP(w, r)
	}
}
