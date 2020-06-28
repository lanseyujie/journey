package router

import (
    "context"
    "errors"
    "fmt"
    "journey/cardinal/log"
    "journey/cardinal/utils"
    "net/http"
    "net/http/httputil"
    "strconv"
    "strings"
    "time"
)

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

// MiddlewareCors
func MiddlewareCors(domain []string) HandlerFunc {
    return func(httpCtx *Context) {
        allow := false
        origin := httpCtx.Input.Header.Get("Origin")
        method := []string{
            http.MethodGet,
            http.MethodPost,
            http.MethodOptions,
            http.MethodPut,
            http.MethodPatch,
            http.MethodDelete,
        }

        if len(domain) > 0 {
            if len(domain) == 1 && domain[0] == "*" {
                allow = true
                origin = "*"
            } else if len(origin) > 0 {
                for _, value := range domain {
                    if strings.Contains(value, origin) {
                        allow = true
                        break
                    }
                }
            }
        }

        if allow {
            httpCtx.Output.Header().Set("Access-Control-Allow-Origin", origin)
            if httpCtx.Input.Method == http.MethodOptions {
                httpCtx.Output.Header().Set("Access-Control-Allow-Methods", strings.Join(method, ", "))
                httpCtx.Output.Header().Set("Access-Control-Allow-Headers", httpCtx.Input.Header.Get("Access-Control-Request-Headers"))
            }
            httpCtx.StatusCode(http.StatusNoContent)
        }
    }
}

// MiddlewareHttpsUpgrade
func MiddlewareHttpsUpgrade(port int) HandlerFunc {
    return func(httpCtx *Context) {
        if httpCtx.Input.TLS == nil {
            url := "https://" + httpCtx.GetHost()
            if port != 443 {
                url += ":" + strconv.FormatInt(int64(port), 10) + httpCtx.Input.URL.Path
            }
            if len(httpCtx.Input.URL.RawQuery) > 0 {
                url += httpCtx.Input.URL.Path + "?" + httpCtx.Input.URL.RawQuery
            } else {
                url += httpCtx.Input.URL.Path
            }
            httpCtx.Redirect(http.StatusMovedPermanently, url)
        } else {
            httpCtx.Next()
        }
    }
}