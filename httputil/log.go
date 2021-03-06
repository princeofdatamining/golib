
package httputil

import (
    "net/http"
    "log"
    "time"
)

type GetLogger func () (*log.Logger)

func NewLogger(logger *log.Logger, getter GetLogger) (Handler) {
    return &Logger{
        logger: logger,
        getter: getter,
    }
}
type Logger struct {
    logger  *log.Logger
    getter  GetLogger
}
func (this *Logger) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
    if this.logger == nil { this.logger = this.getter() }

    start := time.Now()
    host := r.Host
    this.logger.Printf("[%s]Started %s %s", host, r.Method, r.URL.Path)

    next(w, r)

    res := w.(ResponseWriter)
    this.logger.Printf("[%s]Completed %v %s in %v", host, res.Status(), http.StatusText(res.Status()), time.Since(start))
}
