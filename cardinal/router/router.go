package router

import (
    "net/http"
)

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
    router.Group("/").Use(MiddlewareLogger())

    return router
}

// Static will quickly register a static file service route
func (r *Router) Static(prefix, path string) {
    length := len(prefix)
    // make sure to end with /
    if length == 0 || (length > 0 && prefix[length-1] != '/') {
        prefix = prefix + "/"
    }

    r.Get(prefix+":*", func(httpCtx *Context) {
        http.StripPrefix(prefix, http.FileServer(http.Dir(path))).ServeHTTP(httpCtx.Output, httpCtx.Input)
    })
}

// Head
func (r *Router) Head(pattern string, f HandlerFunc) {
    if f != nil {
        r.tree.Insert(http.MethodHead, pattern, f)
    }
}

// Options
func (r *Router) Options(pattern string, f HandlerFunc) {
    if f != nil {
        r.tree.Insert(http.MethodOptions, pattern, f)
    }
}

// Get
func (r *Router) Get(pattern string, f HandlerFunc) {
    if f != nil {
        r.tree.Insert(http.MethodGet, pattern, f)
    }
}

func (r *Router) Post(pattern string, f HandlerFunc) {
    if f != nil {
        r.tree.Insert(http.MethodPost, pattern, f)
    }
}

// Put
func (r *Router) Put(pattern string, f HandlerFunc) {
    if f != nil {
        r.tree.Insert(http.MethodPut, pattern, f)
    }
}

// Delete
func (r *Router) Delete(pattern string, f HandlerFunc) {
    if f != nil {
        r.tree.Insert(http.MethodDelete, pattern, f)
    }
}

// Any
func (r *Router) Any(pattern string, f HandlerFunc) {
    if f != nil {
        r.tree.Insert("ANY", pattern, f)
    }
}

// Insert
func (r *Router) Insert(method, fullRule string, handler HandlerFunc, middleware ...HandlerFunc) {
    r.tree.Insert(method, fullRule, handler, middleware...)
}

// PrintRoutes
func (r *Router) PrintRoutes() {
    r.tree.PrintRoutes(nil)
}

// ServeHTTP
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
    httpCtx := NewContext(rw, req)
    // TODO:// cache
    r.tree.Match(httpCtx, req.URL.Path, req.Method)
    httpCtx.Next()
}
