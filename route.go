package router

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

type Route struct {
	pattern     string
	regexp      *regexp.Regexp
	name        string
	methods     []string
	strictSlash bool
	handler     http.Handler
	middlewares []Middleware
}

type MatchInfo struct {
	route   *Route
	params  url.Values
	handler http.Handler
}

func NewRoute() *Route {
	return &Route{}
}

func (r *Route) Pattern(pattern string) *Route {
	if !strings.HasPrefix(pattern, "/") {
		pattern += "/" + pattern
	}

	buf := make([]byte, 0, len(pattern)+3)
	buf = append(buf, '^')

	if strings.ContainsRune(pattern, '<') && strings.ContainsRune(pattern, '>') {
		re := regexp.MustCompile("<[^>]+>")
		reString := re.ReplaceAllStringFunc(pattern, func(in string) string {
			var converter, name string
			if segments := strings.SplitN(in, ":", 2); len(segments) == 2 {
				last := len(segments[1]) - 1
				converter, name = segments[0][1:], segments[1][:last]
			} else {
				last := len(segments[0]) - 1
				name = segments[0][1:last]
				converter = "any"
			}
			rep := strings.NewReplacer("int", "[\\d]+", "string", "[\\w]+", "any", "[^/]+")
			converter = rep.Replace(converter)
			return fmt.Sprintf("(?P<%s>%s)", name, converter)
		})
		buf = append(buf, reString...)
	} else {
		buf = append(buf, pattern...)
	}

	if strings.HasSuffix(pattern, "/") {
		buf = append(buf, '?')
	} else {
		buf = append(buf, "/?"...)
	}
	buf = append(buf, '$')

	r.pattern = pattern
	r.regexp = regexp.MustCompile(string(buf))
	return r
}

func (r *Route) Handle(handler http.Handler) *Route {
	r.handler = handler
	return r
}

func (r *Route) HandleFunc(f func(http.ResponseWriter, *http.Request)) *Route {
	return r.Handle(http.HandlerFunc(f))
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

func (r *Route) Middlewares(v ...Middleware) *Route {
	for _, middleware := range v {
		r.middlewares = append(r.middlewares, middleware)
	}
	return r
}

func (r *Route) StrictSlash(strictSlash bool) *Route {
	r.strictSlash = strictSlash
	return r
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

func (r *Route) matchPath(path string, match *MatchInfo) bool {
	if out := r.regexp.FindAllStringSubmatch(path, -1); out != nil {
		if r.strictSlash {
			p1 := strings.HasSuffix(r.pattern, "/")
			p2 := strings.HasSuffix(path, "/")

			if p1 != p2 {
				if p1 {
					path += "/"
				} else {
					path = path[:len(path)-1]
				}
				match.handler = http.RedirectHandler(path, http.StatusMovedPermanently)
				return true
			}
		}

		names := r.regexp.SubexpNames()
		for k, v := range out[0] {
			if k > 0 {
				match.params.Add(names[k], v)
			}
		}
		match.handler = r.handler
		return true
	}
	return false
}

func (r *Route) Matches(path, method string) (*MatchInfo, bool) {
	match := &MatchInfo{route: r, params: url.Values{}}
	if r.matchMethod(method) && r.matchPath(path, match) {
		return match, true
	}
	return match, false
}
