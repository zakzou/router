package router

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
)

type Route struct {
	pattern      string
	namedPattern string
	name         string
	strictSlash  bool
	regexp       *regexp.Regexp
	params       map[string]string
	paramNames   []string
	methods      []string
	handler      http.Handler
	middlewares  []MiddlewareInterface
}

func NewRoute() *Route {
	r := new(Route)
	r.params = make(map[string]string)
	return r
}

func (r *Route) Handler(handler http.Handler) *Route {
	r.handler = handler
	return r
}

func (r *Route) Name(name string) *Route {
	r.name = name
	return r
}

func (r *Route) Pattern(pattern string) *Route {
	if ok := strings.HasPrefix(pattern, "/"); !ok {
		pattern = "/" + pattern
	}
	r.pattern = pattern
	return r
}

func (r *Route) Params(params map[string]string) *Route {
	r.params = params
	return r
}

func (r *Route) Param(key, value string) *Route {
	r.params[key] = value
	return r
}

func (r *Route) Methods(v ...string) *Route {
	for _, method := range v {
		r.methods = append(r.methods, strings.ToUpper(method))
	}
	return r
}

func (r *Route) StrictSlash(value bool) *Route {
	r.strictSlash = value
	return r
}

func (r *Route) Middlewares(middlewares ...MiddlewareInterface) *Route {
	for _, middleware := range middlewares {
		r.middlewares = append(r.middlewares, middleware)
	}
	return r
}

func (r *Route) getIndexs() []int {
	var level, idx int
	idxs := make([]int, 0)
	for i := 0; i < len(r.pattern); i++ {
		switch r.pattern[i] {
		case '<':
			if level++; level == 1 {
				idx = i
			}
		case '>':
			if level--; level == 0 {
				if i-idx > 1 {
					idxs = append(idxs, idx, i+1)
				}
			}
		}
	}
	return idxs
}

func (r *Route) parse() {
	if r.regexp != nil {
		return
	}
	defaultPattern := map[string]string{
		"int": "[\\d]+",
		"str": "[\\w]+",
		"any": "[^/]+",
	}
	idxs := r.getIndexs()
	namedPattern, regexpPattern := r.pattern, r.pattern

	for i := 0; i < len(idxs); i += 2 {
		start, end := idxs[i], idxs[i+1]
		origin := r.pattern[start:end]
		target := "[^/]+"
		var paramName string
		if segments := strings.SplitN(r.pattern[start+1:end-1], ":", 2); len(segments) == 2 {
			paramName = segments[1]
			if t, ok := defaultPattern[segments[0]]; ok {
				target = t
			} else {
				target = segments[0]
			}
		} else {
			paramName = segments[0]
		}
		r.paramNames = append(r.paramNames, paramName)
		regexpPattern = strings.Replace(regexpPattern, origin, fmt.Sprintf("(%s)", target), 1)
		namedPattern = strings.Replace(namedPattern, origin, fmt.Sprintf("<%s>", paramName), 1)
	}

	if strings.HasSuffix(regexpPattern, "/") {
		regexpPattern = regexpPattern[:len(regexpPattern)-1]
	}
	regexpPattern += "/?"
	r.regexp = regexp.MustCompile(fmt.Sprintf("^%s$", regexpPattern))
	r.namedPattern = namedPattern
}

func (r *Route) matchMethod(method string) bool {
	method = strings.ToUpper(method)
	for _, v := range r.methods {
		if v == method {
			return true
		}
	}
	if len(r.methods) == 0 && method == "GET" {
		return true
	}
	return false
}

func (r *Route) matchPath(path string) (http.Handler, bool) {
	if r.regexp.MatchString(path) {
		if r.strictSlash {
			p1 := strings.HasSuffix(r.pattern, "/")
			p2 := strings.HasSuffix(path, "/")
			if p1 != p2 {
				if p1 {
					path += "/"
				} else {
					path = path[:len(path)-1]
				}
				return http.RedirectHandler(path, 301), true
			}
		}
		matches := r.regexp.FindStringSubmatch(path)
		if len(r.paramNames) == len(matches)-1 {
			for k, v := range matches[1:] {
				r.Param(r.paramNames[k], v)
			}
		}
		return r.handler, true
	}
	return nil, false
}

func (r *Route) Matches(path, method string) (http.Handler, bool) {
	r.parse()
	if handler, ok := r.matchPath(path); r.matchMethod(method) {
		return handler, ok
	}
	return nil, false
}
