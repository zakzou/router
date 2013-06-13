package router

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	DefaultConverter = "any"

	HttpMethodGet    = "GET"
	HttpMethodPost   = "POST"
	HttpMethodPut    = "PUT"
	HttpMethodDelete = "DELETE"
)

var (
	DefaultReplacer = []string{"int", "\\d+", "string", "[\\w\\-]+", "any", "[^/]+"}
)

type Route struct {
	pattern     string
	regexp      *regexp.Regexp
	handler     http.Handler
	name        string
	methods     []string
	strictSlash bool
	middleware  []http.Handler
}

type MatchedRoute struct {
	route   *Route
	handler http.Handler
	params  url.Values
}

func NewRoute(pattern string, handler http.Handler) *Route {
	route := &Route{pattern: pattern, handler: handler}
	route.parsePattern()
	return route
}

func (r *Route) Name(name string) *Route {
	r.name = name
	return r
}

func (r *Route) Methods(v ...string) *Route {
	for _, method := range v {
		r.methods = append(r.methods, strings.ToUpper(method))
	}
	return r
}

func (r *Route) StrictSlash(strictSlash bool) *Route {
	r.strictSlash = strictSlash
	return r
}

func (r *Route) Middleware(middleware ...http.Handler) *Route {
	r.middleware = append(r.middleware, middleware...)
	return r
}

func (r *Route) parsePattern() {
	if r.regexp != nil {
		return
	}

	pattern := r.pattern
	buf := make([]byte, 0, len(pattern)+3)
	buf = append(buf, '^')

	if strings.ContainsRune(pattern, '<') && strings.ContainsRune(pattern, '>') {
		re := regexp.MustCompile("<[^>]+>")
		reString := re.ReplaceAllStringFunc(pattern, func(in string) string {
			converter := DefaultConverter
			var name string
			if segments := strings.SplitN(in, ":", 2); len(segments) == 2 {
				converter, name = segments[0][1:], segments[1][:len(segments[1])-1]
			} else {
				name = segments[0][1 : len(segments[0])-1]
			}
			return fmt.Sprintf("(?P<%s>%s)", name, strings.NewReplacer(DefaultReplacer...).Replace(converter))
		})
		buf = append(buf, reString...)
	} else {
		buf = append(buf, pattern...)
	}

	if strings.HasSuffix(pattern, "/") {
		buf = append(buf, "?$"...)
	} else {
		buf = append(buf, "/?$"...)
	}
	r.regexp = regexp.MustCompile(string(buf))
}

func (r *Route) supportsHttpMethod(httpMethod string) bool {
	httpMethod = strings.ToUpper(httpMethod)
	if len(r.methods) == 0 {
		if httpMethod == HttpMethodGet {
			return true
		}
	} else {
		for _, method := range r.methods {
			if httpMethod == method {
				return true
			}
		}
	}
	return false
}

func (r *Route) matches(resourceUri string) (MatchedRoute, bool) {
	matched := MatchedRoute{params: url.Values{}}

	if out := r.regexp.FindAllStringSubmatch(resourceUri, -1); out != nil {
		if r.strictSlash {
			p1 := strings.HasSuffix(r.pattern, "/")
			p2 := strings.HasSuffix(resourceUri, "/")
			if p1 != p2 {
				if p1 {
					resourceUri += "/"
				} else {
					resourceUri = resourceUri[:len(resourceUri)-1]
				}
				matched.handler = http.RedirectHandler(resourceUri, http.StatusMovedPermanently)
				return matched, true
			}
		}

		names := r.regexp.SubexpNames()
		for k, v := range out[0] {
			if k > 0 {
				matched.params.Add((names[k]), v)
			}
		}
		matched.handler = r.handler
		return matched, true
	}
	return matched, false
}
