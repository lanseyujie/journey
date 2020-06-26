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
func MiddlewareLogger() MiddlewareFunc {
    return func(handler HandlerFunc) HandlerFunc {
        // create a new handler that includes the last handler
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

            handler(httpCtx)

            // to do something after
        }
    }
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
    t := time.Now()
    httpCtx := NewContext(rw, req)
    httpCtx.since = t
    uri := httpCtx.GetUri()
    method := httpCtx.GetMethod()

    var (
        middleware []MiddlewareFunc
        handler    HandlerFunc
        exist      bool
    )

    // TODO:// cache
    node, params := r.tree.Match(uri)
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

    length := len(middleware)
    // handler push
    for i := range middleware {
        // in reverse order
        m := middleware[length-1-i]
        if m != nil {
            // create a new handler using the current middleware including the last handler
            handler = m(handler)
        }
    }
    // handler pop
    handler(httpCtx)
}
