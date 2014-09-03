
package httputil

import (
    "net/http"
)

type Handler interface {
    ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc)
}

type HandlerProc func (w http.ResponseWriter, r *http.Request, next http.HandlerFunc)

type HandlerFunc HandlerProc
func (this HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
    this(w, r, next)
}

func CallBefore(handler http.Handler) (Handler) {
    return HandlerFunc(func (w http.ResponseWriter, r *http.Request, next http.HandlerFunc) () {
        handler.ServeHTTP(w, r)
        next(w, r)
    })
}
func CallAfter(handler http.Handler) (Handler) {
    return HandlerFunc(func (w http.ResponseWriter, r *http.Request, next http.HandlerFunc) () {
        next(w, r)
        handler.ServeHTTP(w, r)
    })
}

func VoidHandler(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) () {
}

//

type chainNode struct {
    handler Handler
    next    *chainNode
}
func (this *chainNode) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    this.handler.ServeHTTP(w, r, this.next.ServeHTTP)
}

type ChainHandler struct {
    head, tail *chainNode
}
func (this *ChainHandler) Chain(handler Handler) () {
    //* FIFO
    temp := &chainNode{
        handler: handler,
    }
    if this.tail == nil {
        temp.next = this.head
        this.head = temp
    } else {
        temp.next = this.tail.next
        this.tail.next = temp
    }
    this.tail = temp
    //*/
    /* FILO
    this.head = &chainNode{
        handler: handler,
        next: this.head,
    }
    //*/
}
func (this *ChainHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) () {
    w = NewResponseWriter(w)
    this.head.ServeHTTP(w, r)
}
func NewChainHandler(handlers ...Handler) (chain *ChainHandler) {
    chain = &ChainHandler{
        head: &chainNode{
            handler: HandlerFunc(VoidHandler),
        },
    }
    for _, h := range handlers {
        chain.Chain(h)
    }
    return 
}
