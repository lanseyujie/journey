package router

import (
    "context"
    "errors"
    "fmt"
    "journey/cardinal/log"
    "journey/cardinal/utils"
    "net/http"
    "net/http/httputil"
    "time"
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
    r.Get(prefix+"/*", FileServerHandler("/"+prefix+"/", path))
}

// MiddlewareLogger
func MiddlewareLogger() HandlerFunc {
    return func(httpCtx *Context) {
        defer func() {
            var err error
            if e := recover(); e != nil {
                switch e := e.(type) {
                case string:
                    err = errors.New(e)
                case error:
                    err = e
                default:
                    err = errors.New(fmt.Sprint(e))
                }
            }
            if err != nil {
                // print stack trace
                // log.Println(err)
                // debug.PrintStack()
                log.Error(utils.StackTrace(err, 0))

                // dump http request header
                request, _ := httputil.DumpRequest(httpCtx.Input, false)
                log.Debug(string(request))

                // the default error page can be called in the following 3 ways
                httpCtx.Error(http.StatusInternalServerError)
                // GetErrorHandler(http.StatusInternalServerError)(httpCtx)
                // httpCtx.Handler(router.GetErrorHandler(http.StatusInternalServerError))
            }

            // reasons for collecting logs here:
            // 1. capture the response status code, error and running time
            // 2. avoid directly executing the defer process and skip log collection when panic occurs
            log.Http(httpCtx.Logger())
        }()

        // to do something before

        // call the next middleware
        httpCtx.Next()

        // to do something after
    }
}

// MiddlewareTimeout
func MiddlewareTimeout(d time.Duration) HandlerFunc {
    return func(httpCtx *Context) {
        ctx, cancel := context.WithTimeout(httpCtx.Input.Context(), d)
        defer cancel()
        httpCtx.Input = httpCtx.Input.WithContext(ctx)

        httpCtx.Next()
    }
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
