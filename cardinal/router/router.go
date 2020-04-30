package router

import "net/http"

// Router
type Router struct {
    tree *Tree
}

// NewRouter returns a new Router
func NewRouter() *Router {
    return &Router{
        tree: NewTree(),
    }
}

// Default returns a default configured router
func Default() *Router {
    router := NewRouter()

    return router
}

// Static will quickly register a static file service route
func Static(path string) {
    // TODO://
}

// ServeHTTP
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
    httpCtx := NewContext(rw, req)
    // TODO://
}
