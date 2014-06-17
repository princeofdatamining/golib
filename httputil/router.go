
package httputil

import (
    "net/http"
)

type UrlRule struct {
    Method      string
    Pattern     string
    Handler     http.Handler
}
type UrlRules struct {
    builder     map[string]bool
}
func (this *UrlRules) Handle(rules ...*UrlRule) () {
    for _, rule := range rules {
        if _, ok := this.builder[rule.Pattern]; ok {
            continue 
        }
    }
}
func (this *UrlRules) add(method string, rule string, handler http.Handler) () {
    //
}
func (this *UrlRules) parse_rule(rule string) () {
    //
}
