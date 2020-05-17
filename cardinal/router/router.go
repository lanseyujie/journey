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

    return router
}

// Static will quickly register a static file service route
func (r *Router) Static(prefix, path string) {
    r.Get(prefix+"/*", FileServerHandler("/"+prefix+"/", path))
}

func (r *Router) Head(pattern string, f HandlerFunc) {
    if f != nil {
        r.tree.Insert(http.MethodHead, pattern, f)
    }
}

func (r *Router) Options(pattern string, f HandlerFunc) {
    if f != nil {
        r.tree.Insert(http.MethodOptions, pattern, f)
    }
}

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

func (r *Router) Put(pattern string, f HandlerFunc) {
    if f != nil {
        r.tree.Insert(http.MethodPut, pattern, f)
    }
}

func (r *Router) Delete(pattern string, f HandlerFunc) {
    if f != nil {
        r.tree.Insert(http.MethodDelete, pattern, f)
    }
}

func (r *Router) Any(pattern string, f HandlerFunc) {
    if f != nil {
        r.tree.Insert("ANY", pattern, f)
    }
}

func (r *Router) Show() {
    r.tree.Show(nil)
}

// ServeHTTP
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
    httpCtx := NewContext(rw, req)
    uri := httpCtx.GetUri()
    method := httpCtx.GetMethod()
    node, params := r.tree.Match(uri)
    var (
        middleware []MiddlewareFunc
        handler    HandlerFunc
        exist      bool
    )

    if node != nil {
        httpCtx.Input = httpCtx.Input.WithContext(context.WithValue(httpCtx.Input.Context(), "params", params))
        middleware = MiddlewareList(node)
        handler, exist = node.handlers[method]
        if !exist {
            // if the HEAD handler does not exist and the GET handler exists, call the GET handler
            if h, exist := node.handlers[http.MethodGet]; method == http.MethodHead && exist {
                handler = h
            } else {
                // default handler
                handler, _ = node.handlers["ANY"]
            }
        }

        if handler == nil {
            handler = GetErrorHandler(http.StatusMethodNotAllowed)
        }
    } else {
        middleware = r.tree.root.middleware
        handler = GetErrorHandler(http.StatusNotFound)
    }

    if len(middleware) > 0 {
        // TODO:// use chained calls instead it?
        // handler push
        for _, m := range middleware {
            if m != nil {
                handler = m(handler)
            }
        }
    }
    // handler pop
    handler(httpCtx)
}
