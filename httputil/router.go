
package httputil

import (
    "net/http"
    "regexp"
    "strings"
    "fmt"
    "sync"
    "github.com/princeofdatamining/golib/strutil"
)

const (
    MATCH_HOST_ANY = ".*$"
)

var bottle_rule_syntax = regexp.MustCompile(
    `(\\*)`+                                // [1]
    "(?:"+
        "(?:"+
            ":"+
            "([a-zA-Z_][a-zA-Z_0-9]*)?"+    // [2] name
            "()"+                           // [3] filter
            "(?:"+
                "#" + "(.*?)" + "#"+        // [4] conf
            ")"+
        ")"+
        "|"+
        "(?:"+
            "<"+
            "([a-zA-Z_][a-zA-Z_0-9]*)?"+    // [5] name
            "(?:"+
                ":"+
                "([a-zA-Z_]*)"+             // [6] filter
                "(?:"+
                    ":"+
                    "("+                    // [7] conf
                        "(?:"+
                            `\\.`+
                            "|"+
                            `[^\\>]+`+
                        ")+"+
                    ")?"+                   // [7]
                ")?"+
            ")?"+
            ">"+
        ")"+
    ")",
)

func ParseBottleRule(rule string) (parts [][]string) {
    var (
        offset, start, end  int
        prefix, submatch    string
        name, filter, conf  string
    )
    for _, match := range bottle_rule_syntax.FindAllStringSubmatchIndex(rule, -1) {
        // fmt.Printf("% d\n", match)
        m := strutil.SubmatchIndex(match)
        submatch, start, end, _ = m.GetSubmatch(rule, 0)
        prefix += rule[offset:start]
        // fmt.Printf("prefix +=> %q\n", prefix)
        offset = end
        //
        start, end, _ = m.GetIndexPairs(1)
        if (end - start) % 2 != 0 {
            prefix += submatch[end-start:]
            // fmt.Printf("prefix +[\\:]=> %q\n", prefix)
            continue 
        }
        if prefix != "" {
            parts = append(parts, []string{prefix, "", "", ""})
            // fmt.Printf("prefix => %q\n", prefix)
            prefix = ""
        }
        //
        if _, _, _, exists := m.GetSubmatch(rule, 3); exists {
            name   = m.GetSubstring(rule, 2)
            filter = m.GetSubstring(rule, 3)
            conf   = m.GetSubstring(rule, 4)
        } else {
            name   = m.GetSubstring(rule, 5)
            filter = m.GetSubstring(rule, 6)
            conf   = m.GetSubstring(rule, 7)
        }
        // fmt.Printf("name: %q, filter: %q, conf: %q\n", name, filter, conf)
        parts = append(parts, []string{"", name, filter, conf})
    }
    if offset < len(rule) || prefix != "" {
        prefix += rule[offset:]
        // fmt.Printf("prefix ::=> %q\n", prefix)
        parts = append(parts, []string{prefix, "", "", ""})
    }
    return 
}

//

type WrapperFunc func (http.Handler) (http.Handler)

//

func ErrorHandler(error string, code int) (http.Handler) { return &errorHandler{error, code} }
type errorHandler struct {
    error   string
    code    int
}
func (this *errorHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) () { http.Error(w, this.error, this.code) }

//

func FileHandler(fn string) (http.Handler) { return &fileHandler{fn} }
type fileHandler struct {
    fn    string
}
func (this *fileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) () { http.ServeFile(w, r, this.fn) }

//

var _ http.Handler = NewMultiHostRouter()

func NewMultiHostRouter() (*MultiHostRouter) {
    return &MultiHostRouter{
        cached: make(map[string]*Router),
        DefaultSchema: "http",
    }
}
type MultiHostRouter struct {
    sync.RWMutex
    hosts           []*Router
    defaults        *Router
    cached          map[string]*Router
    //
    wrapFunc        WrapperFunc
    DefaultSchema   string
    DefaultHost     string
    UseDefaultHost  bool
}
func (this *MultiHostRouter) ServeHTTP(w http.ResponseWriter, r *http.Request) () {
    h := this.Handler(r)
    h.ServeHTTP(w, r)
}
func (this *MultiHostRouter) Handler(r *http.Request) (h http.Handler) {
    this.RLock()
    defer this.RUnlock()

    routers, found := this.findAllRouters(r.Host)
    if !found {
        if this.UseDefaultHost && this.DefaultHost != r.Host {
            url := *r.URL
            url.Scheme = this.DefaultSchema
            url.Host = this.DefaultHost
            return this.wrapped(http.RedirectHandler(url.String(), http.StatusTemporaryRedirect))
        } else {
            return this.wrapped(ErrorHandler("406 Unknown Host", http.StatusNotAcceptable))
        }
    }
    for _, router := range routers {
        if route, _, _ := router.match(r.URL.Path); route != nil {
            return route.handler
        }
    }
    return this.wrapped(http.NotFoundHandler())
}
func (this *MultiHostRouter) findAllRouters(host string) (routers []*Router, found bool) {
    routers, found = this.findRouters(host)
    if found || !this.UseDefaultHost {
        return 
    }
    return this.findRouters(this.DefaultHost)
}
func (this *MultiHostRouter) findRouters(host string) (routers []*Router, found bool) {
    host = strings.Split(host, ":")[0]
    for _, router := range this.hosts {
        if router.matchHost(host) {
            routers, found = append(routers, router), true
        }
    }
    if router := this.defaults; !found && router != nil {
        if router.matchHost(host) {
            routers, found = append(routers, router), true
        }
    }
    return 
}
func (this *MultiHostRouter) WrapFunc(f WrapperFunc) () { this.wrapFunc = f }
func (this *MultiHostRouter) wrapped(h http.Handler) (http.Handler) {
    if this.wrapFunc == nil {
        return h
    }
    return this.wrapFunc(h)
}
func (this *MultiHostRouter) Handle(host_pattern, path_pattern string, h http.Handler) () { this.AddRouter(host_pattern).Handle(path_pattern, h) }
func (this *MultiHostRouter) cacheRouter(host_pattern string) (router *Router, found bool) {
    if this.cached == nil {
        this.cached = make(map[string]*Router)
    }
    router, found = this.cached[host_pattern]
    return 
}
func (this *MultiHostRouter) AddRouter(host_pattern string) (router *Router) {
    this.Lock()
    if router, found := this.cacheRouter(host_pattern); found {
        this.Unlock()
        return router
    }
    defer this.Unlock()

    router = newRouter(host_pattern)
    router.WrapFunc(this.wrapFunc)
    //
    if host_pattern == MATCH_HOST_ANY {
        this.defaults = router
    } else {
        this.hosts = append(this.hosts, router)
    }
    this.cached[host_pattern] = router
    return 
}

//

type Router struct {
    sync.RWMutex
    rules       map[string]*Route
    builder     map[string][]*builderPart
    static      map[string]*Route
    //
    host_re     *regexp.Regexp
    wrapFunc    WrapperFunc
}
type builderPart struct {
    key     string
    static  bool
}
func newRouter(host_pattern string) (*Router) {
    return &Router{
        host_re: regexp.MustCompile(host_pattern),
        rules: make(map[string]*Route),
        builder: make(map[string][]*builderPart),
        static: make(map[string]*Route),
    }
}
func (this *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) () { this.Handler(r).ServeHTTP(w, r) }
func (this *Router) Handler(r *http.Request) (http.Handler) {
    if route, _, _ := this.match(r.URL.Path); route != nil {
        return route.handler
    }
    return this.wrapped(http.NotFoundHandler())
}
func (this *Router) matchHost(host string) (bool) { return this.host_re.MatchString(host) }
func (this *Router) match(path string) (*Route, []string, map[string]string) {
    this.RLock()
    defer this.RUnlock()

    if rule, found := this.static[path]; found {
        return rule, nil, nil
    }
    return nil, nil, nil
}
func (this *Router) WrapFunc(f WrapperFunc) () { this.wrapFunc = f }
func (this *Router) wrapped(h http.Handler) (http.Handler) {
    if this.wrapFunc == nil {
        return h
    }
    return this.wrapFunc(h)
}
func (this *Router) Handle(rule string, h http.Handler) () {
    this.Lock()
    defer this.Unlock()

    if _, exists := this.rules[rule]; exists {
        return 
    }

    route := &Route{
        handler: this.wrapped(h),
    }
    this.rules[rule] = route

    pattern := ""
    builder := []*builderPart{}
    static := true
    // anons := 0
    for _, parts := range ParseBottleRule(rule) {
        prefix, key, filter, conf := parts[0], parts[1], parts[2], parts[3]
        fmt.Printf("\t%q %q %q %q\n", prefix, key, filter, conf)
        if prefix == "" {
            static = false
            _, _, _ = key, filter, conf
        } else {
            pattern += regexp.QuoteMeta(prefix)
            builder = append(builder, &builderPart{ prefix, true })
        }
    }
    this.builder[rule] = builder

    if static {
        path := this.build(rule)
        fmt.Printf("\tstatic %q\n", path)
        this.static[path] = route
        return 
    }
}
func (this *Router) build(rule string) (url string) {
    builder := this.builder[rule]
    for _, part := range builder {
        if part.static {
            url += part.key
        }
    }
    return 
}

//

type Route struct {
    handler     http.Handler
}
