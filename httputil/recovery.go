
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

    host := r.Host
    defer func() {
        if err := recover(); err != nil {
            w.WriteHeader(http.StatusInternalServerError)
            stack := debug.Stack()
            f := "[%s]PANIC: %s\n%s"
            msg := fmt.Sprintf(f, host, err, stack)
            this.logger.Print(msg)

            if this.PrintStack {
                fmt.Fprint(w, msg)
            }
        }
    }()

    next(w, r)
}
