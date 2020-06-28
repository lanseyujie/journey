package router

import (
    "context"
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
    r.Get(prefix+"*", func(httpCtx *Context) {
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

// Show
func (r *Router) Show() {
    r.tree.Show(nil)
}

// ServeHTTP
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
    httpCtx := NewContext(rw, req)
    uri := httpCtx.GetUri()
    method := httpCtx.GetMethod()

    var exist bool
    // TODO:// cache
    node, params := r.tree.Match(uri)
    if node != nil {
        httpCtx.Input = httpCtx.Input.WithContext(context.WithValue(httpCtx.Input.Context(), "params", params))
        httpCtx.handler, exist = node.handlers[method]
        if !exist {
            // if the HEAD handler does not exist and the GET handler exists, call the GET handler
            if h, exist := node.handlers[http.MethodGet]; method == http.MethodHead && exist {
                httpCtx.handler = h
            } else {
                // default handler
                httpCtx.handler, _ = node.handlers["ANY"]
            }
        }

        if httpCtx.handler == nil {
            httpCtx.handler = GetErrorHandler(http.StatusMethodNotAllowed)
        } else {
            httpCtx.middleware = MiddlewareList(node)
        }
    } else {
        httpCtx.handler = GetErrorHandler(http.StatusNotFound)
    }

    httpCtx.Next()
}
