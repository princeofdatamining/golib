
package httputil

import (
    "net/http"
    "regexp"
    "strings"
    "fmt"
    "sync"
    "github.com/princeofdatamining/golib/strutil"
)

/*
    <name>
    <name:filter>
    <name:re:pattern>
//*/
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

type FilterFunc func (conf string) (string)

var filters = map[string]FilterFunc{
    "re": func (conf string) (string) {
        if conf != "" { return conf }
        return `[^/]+`
    },
    "int": func (conf string) (string) {
        return `-?\d+`
    },
    "float": func (conf string) (string) {
        return `-?[\d.]+`
    },
    "path": func (conf string) (string) {
        return `.+`
    },
}

//

type (
    WrapperFunc func (http.Handler) (http.Handler)
    ArgsHandler func (http.ResponseWriter, *http.Request, []string, map[string]string) ()
)

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

const (
    MATCH_HOST_ANY = ".*$"
    METHOD_ANY = "ANY"
)

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
        if h = router.match(r.URL.Path, r.Method); h != nil {
            return h
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
func (this *MultiHostRouter) Handle     (host_pattern, path_pattern, method string, h http.Handler) () {
    this.AddRouter(host_pattern).Handle     (path_pattern, method, h)
}
func (this *MultiHostRouter) Handles    (host_pattern, path_pattern string, methods map[string]http.Handler) () {
    this.AddRouter(host_pattern).Handles    (path_pattern, methods  )
}
func (this *MultiHostRouter) HandleFunc (host_pattern, path_pattern, method string, f ArgsHandler ) () {
    this.AddRouter(host_pattern).HandleFunc (path_pattern, method, f)
}
func (this *MultiHostRouter) HandleFuncs(host_pattern, path_pattern string, methods map[string]ArgsHandler ) () {
    this.AddRouter(host_pattern).HandleFuncs(path_pattern, methods  )
}
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

    router = newRouter(this, host_pattern)
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

func NewRouter() (*Router) {
    return &Router{
        rules: make(map[string]map[string]*Route),
        builder: make(map[string][]*builderPart),
        static: make(map[string]map[string]*Route),
    }
}
type Router struct {
    sync.RWMutex
    rules       map[string]map[string]*Route
    builder     map[string][]*builderPart
    static      map[string]map[string]*Route
    dynamic     []*dynamicPart
    //
    parent      *MultiHostRouter
    host_re     *regexp.Regexp
    wrapFunc    WrapperFunc
}
type builderPart struct {
    key     string
    static  bool
}
type dynamicPart struct {
    pattern     string
    rexp        *regexp.Regexp
    pairs       []*pairGetargsRule
}
type pairGetargsRule struct {
    getargs     func (string) ([]string, map[string]string)
    targets     map[string]*Route
}
func newRouter(p *MultiHostRouter, host_pattern string) (*Router) {
    this := NewRouter()
    this.parent = p
    this.host_re = regexp.MustCompile(host_pattern)
    this.WrapFunc(p.wrapFunc)
    return this
}
func (this *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) () { this.Handler(r).ServeHTTP(w, r) }
func (this *Router) Handler(r *http.Request) (h http.Handler) {
    if h = this.match(r.URL.Path, r.Method); h != nil {
        return h
    }
    return this.wrapped(http.NotFoundHandler())
}
func (this *Router) matchHost(host string) (bool) { return this.host_re.MatchString(host) }
func getRouteByMethod(targets map[string]*Route, method string) (route *Route) {
    var ok bool
    if route, ok = targets[method]; ok {
        return 
    }
    if route, ok = targets[METHOD_ANY]; ok {
        return 
    }
    return nil
}
func (this *Router) match(path, method string) (h http.Handler) {
    this.RLock()
    defer this.RUnlock()

    // fmt.Printf("match static rule...\n")
    if targets, found := this.static[path]; found {
        route := getRouteByMethod(targets, method)
        return route.makeHandler(nil, nil, nil)
    }

    // fmt.Printf("match dynamic rule...\n")
    for _, d := range this.dynamic {
        // fmt.Printf("dynamic %q\n", d.pattern)
        subindex := d.rexp.FindStringSubmatchIndex(path)
        // fmt.Printf("\t% d\n", subindex)
        i := indexCombined(subindex, len(d.pairs))-1
        if i < 0 {
            continue
        }
        getargs, targets := d.pairs[i].getargs, d.pairs[i].targets
        args, kwargs := getargs(path)
        // fmt.Printf("\tmatched args: % q; kwargs: %+v\n", args, kwargs)
        route := getRouteByMethod(targets, method)
        return route.makeHandler(args, kwargs, this.wrapped)
    }

    return nil
}
func indexCombined(subs []int, n int) (i int) {
    for i < n {
        i++
        if subs[i*2+1] >= 0 {
            return i
        }
    }
    return 0
}
func (this *Router) WrapFunc(f WrapperFunc) () { this.wrapFunc = f }
func (this *Router) wrapped(h http.Handler) (http.Handler) {
    if this.wrapFunc == nil {
        return h
    }
    return this.wrapFunc(h)
}
func (this *Router) handle(rule, method string, route *Route) () {
    this.Lock()
    defer this.Unlock()

    if method == "*" || method == "" {
        method = METHOD_ANY
    }
    if ruleMethods, exists := this.rules[rule]; exists {
        ruleMethods[method] = route
        return 
    }
    targets := map[string]*Route{method:route}
    this.rules[rule] = targets

    pattern := ""
    flat_pattern := ""
    builder := []*builderPart{}
    static := true
    anons := 0
    for _, parts := range ParseBottleRule(rule) {
        var se string
        prefix, key, filter, conf := parts[0], parts[1], parts[2], parts[3]
        // fmt.Printf("\t%q %q %q %q\n", prefix, key, filter, conf)
        if prefix != "" {
            se = regexp.QuoteMeta(prefix)
            pattern += se
            flat_pattern += se
            builder = append(builder, &builderPart{ prefix, true })
            continue
        }
        static = false
        mask := filters[filter](conf)
        if key == "" {
            // key = fmt.Sprintf("anon%d", anons)
            pattern += fmt.Sprintf("(%s)", mask)
            anons++
        } else {
            pattern += fmt.Sprintf("(?P<%s>%s)", key, mask)
        }
        flat_pattern += fmt.Sprintf("(?:%s)", mask)
        builder = append(builder, &builderPart{ key, false })
    }
    this.builder[rule] = builder

    if static {
        path := this.build(rule)
        // fmt.Printf("\tstatic %q\n", path)
        this.static[path] = targets
        return 
    }

    reMatch := regexp.MustCompile(fmt.Sprintf("^%s$", pattern))
    subNames := reMatch.SubexpNames()
    getargs := func (path string) (args []string, kwargs map[string]string) {
        subs := reMatch.FindStringSubmatch(path)
        for i, subName := range subNames {
            if i == 0 { continue }
            if subName == "" {
                args = append(args, subs[i])
            } else {
                if kwargs == nil {
                    kwargs = make(map[string]string)
                }
                kwargs[subName] = subs[i]
            }
        }
        return args, kwargs
    }

    var (
        e error
        last *dynamicPart
    )
    flat_pattern = fmt.Sprintf("(^%s$)", flat_pattern)
    if N := len(this.dynamic); N <= 0 {
        e = fmt.Errorf("")
    } else {
        last = this.dynamic[N-1]
        combined := last.pattern + "|" + flat_pattern
        if exp, err := regexp.Compile(combined); err == nil {
            last.pattern = combined
            last.rexp = exp
        } else {
            e = err
        }
    }
    if e != nil {
        exp, err := regexp.Compile(flat_pattern)
        if err != nil {
            return 
        }
        last = &dynamicPart{
            pattern: flat_pattern,
            rexp: exp,
        }
        this.dynamic = append(this.dynamic, last)
    }
    last.pairs = append(last.pairs, &pairGetargsRule{
        getargs: getargs,
        targets: targets,
    })
}
func (this *Router) Handle     (rule, method string, h http.Handler) () {
    this.handle(rule, method, &Route{
        nullArgsHandler: this.wrapped(h),
    })
}
func (this *Router) Handles    (rule string, methods map[string]http.Handler) () {
    for method, h := range methods {
        this.Handle(rule, method, h)
    }
}
func (this *Router) HandleFunc (rule, method string, f ArgsHandler ) () {
    this.handle(rule, method, &Route{
        withArgsHandler: f,
    })
}
func (this *Router) HandleFuncs(rule string, methods map[string]ArgsHandler ) () {
    for method, f := range methods {
        this.HandleFunc(rule, method, f)
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
    nullArgsHandler     http.Handler
    withArgsHandler     ArgsHandler
}
func (this *Route) makeHandler(args []string, kwargs map[string]string, wrapped WrapperFunc) (http.Handler) {
    if this == nil {
        return nil
    }
    if this.nullArgsHandler != nil {
        return this.nullArgsHandler
    }
    if this.withArgsHandler != nil {
        return wrapped(&argsHandler{
            handler : this.withArgsHandler,
            args    : args,
            kwargs  : kwargs,
        })
    }
    return nil
}

type argsHandler struct {
    handler ArgsHandler
    args    []string
    kwargs  map[string]string
}
func (this *argsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) () {
    this.handler(w, r, this.args, this.kwargs)
}

//
