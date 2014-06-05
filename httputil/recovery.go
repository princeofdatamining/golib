
package httputil

import (
    "net/http"
    "log"
    "fmt"
    "runtime/debug"
)

func NewRecovery(logger *log.Logger, getter GetLogger, debug bool) (Handler) {
    return &Recovery{
        logger: logger,
        getter: getter,
        PrintStack: debug,
    }
}
type Recovery struct {
    logger  *log.Logger
    getter  GetLogger
    PrintStack bool
}
func (this *Recovery) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
    if this.logger == nil { this.logger = this.getter() }

    defer func() {
        if err := recover(); err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            stack := debug.Stack()
            f := "PANIC: %s\n%s"
            this.logger.Printf(f, err, stack)

            if this.PrintStack {
                fmt.Fprintf(w, f, err, stack)
            }
        }
    }()

    next(w, r)
}
