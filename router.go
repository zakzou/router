package router

import (
	"fmt"
	"net/http"
	"regexp"
)

type Router struct {
	namedRoutes     map[string]*Route
	routes          []*Route
	NotFoundHandler http.Handler
	strictSlash     bool
	middlewares     []Middleware
	hooks           map[string][]Hook
}

func NewRouter() *Router {
	return &Router{namedRoutes: make(map[string]*Route), hooks: make(map[string][]Hook)}
}

func (r *Router) Handle(pattern string, handler http.Handler) *Route {
	route := NewRoute().Handle(handler).Pattern(pattern)
	route.StrictSlash(r.strictSlash).Middlewares(r.middlewares...)
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

func (r *Router) Middlewares(middlewares ...Middleware) *Router {
	for _, middleware := range middlewares {
		r.middlewares = append(r.middlewares, middleware)
	}
	return r
}

func (r *Router) MiddlewareFunc(f func(http.ResponseWriter, *http.Request)) *Router {
	return r.Middlewares(MiddlewareFunc(f))
}

func (r *Router) Hooks(name string, hooks ...Hook) *Router {
	r.hooks[name] = append(r.hooks[name], hooks...)
	return r
}

func (r *Router) HookFunc(name string, f func(http.ResponseWriter, *http.Request)) *Router {
	return r.Hooks(name, HookFunc(f))
}

func (r *Router) ApplyHook(name string, rw http.ResponseWriter, req *http.Request) {
	if hooks, ok := r.hooks[name]; ok {
		for _, hook := range hooks {
			hook.ServeHTTP(rw, req)
		}
	}
}

func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	r.ApplyHook(HookBefore, rw, req)
	defer r.ApplyHook(HookAfter, rw, req)

	requestPath, requestMethod := req.URL.Path, req.Method
	var handler http.Handler

	for _, route := range r.routes {
		if matched, ok := route.Matches(requestPath, requestMethod); ok {
			handler = matched.handler
			for _, middleware := range matched.route.middlewares {
				middleware.ServeHTTP(rw, req)
			}
			if rawEncode := matched.params.Encode(); rawEncode != "" {
				req.URL.RawQuery = rawEncode + "&" + req.URL.RawQuery
			}
			break
		}
	}

	if handler == nil {
		if r.NotFoundHandler != nil {
			handler = r.NotFoundHandler
		} else {
			handler = http.NotFoundHandler()
		}
	}
	handler.ServeHTTP(rw, req)
}

func (r *Router) parseNamedRoute() {
	if r.namedRoutes == nil {
		for _, route := range r.routes {
			if route.name != "" {
				r.namedRoutes[route.name] = route
			}
		}
	}
}

func (r *Router) UrlFor(name string, params map[string]interface{}) (string, bool) {
	r.parseNamedRoute()
	if route, ok := r.namedRoutes[name]; ok {
		url := route.pattern
		for key, value := range params {
			url = regexp.MustCompile(fmt.Sprintf("<[^:/>]*:?%s>", key)).ReplaceAllString(url, fmt.Sprintf("%v", value))
		}
		return url, true
	}
	return "", false
}
