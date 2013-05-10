package router

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type Router struct {
	prefix          string
	namedRoutes     map[string]*Route
	parse           bool
	routes          []*Route
	strictSlash     bool
	NotFoundHandler http.Handler
	middlewares     []MiddlewareInterface
}

func NewRouter() *Router {
	r := new(Router)
	r.namedRoutes = make(map[string]*Route)
	return r
}

func (r *Router) StrictSlash(value bool) *Router {
	r.strictSlash = value
	return r
}

func (r *Router) SubRouter() *Router {
	r.prefix = ""
	r.strictSlash = false
	r.middlewares = make([]MiddlewareInterface, 0)
	return r
}

func (r *Router) Prefix(prefix string) *Router {
	if ok := strings.HasPrefix(prefix, "/"); !ok {
		prefix = "/" + prefix
	}
	if ok := strings.HasSuffix(prefix, "/"); ok {
		prefix = prefix[:len(prefix)-1]
	}
	r.prefix = prefix
	return r
}

func (r *Router) Middlewares(middlewares ...MiddlewareInterface) *Router {
	for _, middleware := range middlewares {
		r.middlewares = append(r.middlewares, middleware)
	}
	return r
}

func (r *Router) Handle(pattern string, handler http.Handler) *Route {
	pattern = r.prefix + pattern
	route := NewRoute().Pattern(pattern).Handler(handler)
	route.StrictSlash(r.strictSlash).Middlewares(r.middlewares...)
	r.routes = append(r.routes, route)
	return route
}

func (r *Router) HandleFunc(pattern string, f func(http.ResponseWriter, *http.Request)) *Route {
	return r.Handle(pattern, http.HandlerFunc(f))
}

func (this *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Path
	requestMethod := r.Method
	var handler http.Handler
	var ok bool

	for _, route := range this.routes {
		if handler, ok = route.Matches(requestPath, requestMethod); ok {
			values := url.Values{}
			for k, v := range route.params {
				values.Add(k, v)
			}
			r.URL.RawQuery = values.Encode() + "&" + r.URL.RawQuery
			for _, middleware := range route.middlewares {
				middleware.BeforeRequest(w, r)
			}
			handler.ServeHTTP(w, r)
			for _, middleware := range route.middlewares {
				middleware.AfterRequest(w, r)
			}
			break
		}
	}

	if handler == nil {
		if this.NotFoundHandler != nil {
			handler = this.NotFoundHandler
		} else {
			handler = http.NotFoundHandler()
		}
		handler.ServeHTTP(w, r)
	}
}

func (r *Router) parseNamedRoute() {
	if r.parse {
		return
	}
	for _, route := range r.routes {
		if route.name != "" {
			r.namedRoutes[route.name] = route
		}
	}
	r.parse = true
}

func (r *Router) UrlFor(name string, params map[string]interface{}) string {
	r.parseNamedRoute()
	var namedPattern string
	if route, ok := r.namedRoutes[name]; ok {
		namedPattern = route.namedPattern
		for key, value := range params {
			valueString := fmt.Sprintf("%v", value)
			namedPattern = strings.Replace(namedPattern, fmt.Sprintf("<%s>", key), valueString, -1)
		}
	}
	return namedPattern
}
