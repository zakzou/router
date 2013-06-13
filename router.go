package router

import (
	"fmt"
	"net/http"
	"regexp"
)

const (
	HookBeforeRouter   = "hook.before.router"
	HookBeforeDispatch = "hook.before.dispatch"
	HookAfterRouter    = "hook.after.router"
	HookAfterDispatch  = "hook.after.dispatch"
)

type Router struct {
	namedRoutes     map[string]*Route
	routes          []*Route
	hooks           map[string][]http.Handler
	middleware      []http.Handler
	strictSlash     bool
	NotFoundHandler http.Handler
}

func NewRouter() *Router {
	return &Router{namedRoutes: make(map[string]*Route), hooks: make(map[string][]http.Handler)}
}

func (r *Router) Handle(pattern string, handler http.Handler) *Route {
	route := NewRoute(pattern, handler)
	route.StrictSlash(r.strictSlash).Middleware(r.middleware...)
	r.routes = append(r.routes, route)
	return route
}

func (r *Router) HandleFunc(pattern string, f func(http.ResponseWriter, *http.Request)) *Route {
	return r.Handle(pattern, http.HandlerFunc(f))
}

func (r *Router) StrictSlash(strictSlash bool) *Router {
	r.strictSlash = strictSlash
	return r
}

func (r *Router) MiddlewareFunc(f func(http.ResponseWriter, *http.Request)) *Router {
	r.middleware = append(r.middleware, http.HandlerFunc(f))
	return r
}

func (r *Router) HookFunc(name string, f func(http.ResponseWriter, *http.Request)) *Router {
	r.hooks[name] = append(r.hooks[name], http.HandlerFunc(f))
	return r
}

func (r *Router) ApplyHook(name string, rw http.ResponseWriter, req *http.Request) {
	if hooks, ok := r.hooks[name]; ok {
		for _, hook := range hooks {
			hook.ServeHTTP(rw, req)
		}
	}
}

func (r *Router) getMatchedRoutes(httpMethod, resourceUri string) []MatchedRoute {
	var matchedRoutes = make([]MatchedRoute, 0)
	for _, route := range r.routes {
		if route.supportsHttpMethod(httpMethod) {
			if matched, ok := route.matches(resourceUri); ok {
				matchedRoutes = append(matchedRoutes, matched)
			}
		}
	}
	return matchedRoutes
}

func (r *Router) dispatch(matched MatchedRoute, rw http.ResponseWriter, req *http.Request) bool {
	for _, m := range matched.route.middleware {
		m.ServeHTTP(rw, req)
	}

	if rawEncode := matched.params.Encode(); rawEncode != "" {
		req.URL.RawQuery = rawEncode + "&" + req.URL.RawQuery
	}

	matched.handler.ServeHTTP(rw, req)
	return true
}

func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	r.ApplyHook(HookBeforeRouter, rw, req)
	defer r.ApplyHook(HookAfterRouter, rw, req)

	var dispatched bool
	if matchedRoutes := r.getMatchedRoutes(req.Method, req.URL.Path); len(matchedRoutes) > 0 {
		for _, matched := range matchedRoutes {
			r.ApplyHook(HookBeforeDispatch, rw, req)
			if dispatched = r.dispatch(matched, rw, req); dispatched {
				break
			}
			r.ApplyHook(HookAfterDispatch, rw, req)
		}
	}

	if !dispatched {
		if r.NotFoundHandler != nil {
			r.NotFoundHandler.ServeHTTP(rw, req)
		} else {
			http.NotFoundHandler().ServeHTTP(rw, req)
		}
	}
}

func (r *Router) parseNameRoutes() {
	if r.namedRoutes == nil {
		for _, route := range r.routes {
			if route.name != "" {
				r.namedRoutes[route.name] = route
			}
		}
	}
}

func (r *Router) UrlFor(name string, params map[string]interface{}) (string, bool) {
	r.parseNameRoutes()
	if route, ok := r.namedRoutes[name]; ok {
		url := route.pattern
		for key, value := range params {
			url = regexp.MustCompile(fmt.Sprintf("<[^:/>]*:?%s>", key)).ReplaceAllString(url, fmt.Sprintf("%v", value))
		}
		return url, true
	}
	return "", false
}
