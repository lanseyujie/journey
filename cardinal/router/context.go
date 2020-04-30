package router

import (
    "net/http"
    "strconv"
    "strings"
)

// Context is the router context
type Context struct {
    Input  *http.Request
    Output http.ResponseWriter
}

// NewContext returns a new router context
func NewContext(rw http.ResponseWriter, req *http.Request) *Context {
    return &Context{
        Input:  req,
        Output: rw,
    }
}

// GetHost returns request host
func (ctx *Context) GetHost() string {
    return strings.Split(ctx.Input.Host, ":")[0]
}

// GetPort returns request port
func (ctx *Context) GetPort() (port int) {
    if socket := strings.Split(ctx.Input.Host, ":"); len(socket) == 2 {
        port, _ = strconv.Atoi(socket[1])
    } else if ctx.Input.TLS != nil {
        port = 443
    } else {
        port = 80
    }

    return
}

// GetMethod returns request method
func (ctx *Context) GetMethod() string {
    return ctx.Input.Method
}

// GetScheme returns request scheme
func (ctx *Context) GetScheme() string {
    return ctx.Input.URL.Scheme
}

// GetUri returns request uri
func (ctx *Context) GetUri() string {
    return ctx.Input.URL.Path
}

// GetQuery returns a GET request parameter
func (ctx *Context) GetQuery(key string) string {
    return ctx.Input.URL.Query().Get(key)
}
