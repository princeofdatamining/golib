
package httputil

import (
    "net/http"
    "net"
    "bufio"
    "errors"
)

type BeforeFunc func (ResponseWriter)

type ResponseWriter interface {
    http.ResponseWriter
    http.Flusher
    Status() (int)
    ContentType() (string)
    SetContentType(string) ()
    Written() (bool)
    Size()  (int)
    Before(BeforeFunc) ()
}

func NewResponseWriter(rw http.ResponseWriter) (ResponseWriter) {
    return &responseWriter{
        ResponseWriter: rw,
    }
}
type responseWriter struct {
    http.ResponseWriter
    status          int
    contentType     string
    size            int
    beforeFuncs     []BeforeFunc
}
func (this *responseWriter) Status() (int) { return this.status }
func (this *responseWriter) Written() (bool) { return this.status != 0 }
func (this *responseWriter) ContentType() (string) { return this.contentType }
func (this *responseWriter) Size() (int) { return this.size }
func (this *responseWriter) Before(before BeforeFunc) { this.beforeFuncs = append(this.beforeFuncs, before) }

func (this *responseWriter) Flush() () {
    if flusher, ok := this.ResponseWriter.(http.Flusher); ok { flusher.Flush() }
}
func (this *responseWriter) CloseNotify() <-chan bool {
    return this.ResponseWriter.(http.CloseNotifier).CloseNotify()
}
func (this *responseWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
    if hijacker, ok := this.ResponseWriter.(http.Hijacker); ok { return hijacker.Hijack() }
    return nil, nil, ErrNotHijacker
}
var ErrNotHijacker = errors.New("the ResponseWriter doesn't support the Hijacker interface")

func (this *responseWriter) callBefore() () {
    for i := len(this.beforeFuncs)-1; i >= 0; i-- { this.beforeFuncs[i](this) }
}
func (this *responseWriter) SetContentType(t string) () {
    this.contentType = t
}
func (this *responseWriter) WriteHeader(status int) {
    this.callBefore()
    this.ResponseWriter.WriteHeader(status)
    this.status = status
    this.Header().Set(HEADER_CONTENT_TYPE, this.contentType)
}
func (this *responseWriter) Write(buf []byte) (size int, err error) {
    if !this.Written() {
        this.WriteHeader(http.StatusOK)
    }
    size, err = this.ResponseWriter.Write(buf)
    this.size += size
    return 
}
