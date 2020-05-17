package router

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net"
    "net/http"
    "strconv"
    "strings"
    "time"
)

var errorHandler = make(map[int]HandlerFunc)

type Array map[string]interface{}

// Context is the router context
type Context struct {
    Input  *http.Request
    Output http.ResponseWriter
    code   int
}

// NewContext returns a new router context
func NewContext(rw http.ResponseWriter, req *http.Request) *Context {
    return &Context{
        Input:  req,
        Output: rw,
        code:   http.StatusOK,
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

// GetParams
func (ctx *Context) GetParams() map[string]string {
    if params, ok := ctx.Input.Context().Value("params").(map[string]string); ok {
        return params
    }

    return nil
}

// GetHeader
func (ctx *Context) GetHeader(key string) string {
    return ctx.Input.Header.Get(key)
}

// GetAuth
func (ctx *Context) GetAuth() string {
    return ctx.Input.Header.Get("Authentication")
}

// GetReferer
func (ctx *Context) GetReferer() string {
    return ctx.Input.Header.Get("Referer")
}

// GetUserAgent
func (ctx *Context) GetUserAgent() string {
    return ctx.Input.UserAgent()
}

// GetClientIp
func (ctx *Context) GetClientIp() string {
    ip := strings.TrimSpace(strings.Split(ctx.Input.Header.Get("X-Forwarded-For"), ",")[0])
    if ip != "" {
        return ip
    }

    ip = strings.TrimSpace(ctx.Input.Header.Get("X-Real-Ip"))
    if ip != "" {
        return ip
    }

    if ip, _, err := net.SplitHostPort(strings.TrimSpace(ctx.Input.RemoteAddr)); err == nil {
        return ip
    }

    return ""
}

// GetQuery returns a GET request parameter
func (ctx *Context) GetQuery(key string) string {
    return ctx.Input.URL.Query().Get(key)
}

// GetPostFrom returns a Post form value
func (ctx *Context) GetPostFrom(key string) string {
    return ctx.Input.PostFormValue(key)
}

// GetBody
func (ctx *Context) GetBody() ([]byte, error) {
    return ioutil.ReadAll(ctx.Input.Body)
}

// GetJson and parse it
func (ctx *Context) GetJson(m interface{}) (interface{}, error) {
    ret, err := ioutil.ReadAll(ctx.Input.Body)
    if err != nil {
        return m, err
    }

    err = json.Unmarshal(ret, &m)
    if err != nil {
        return m, err
    }

    return m, nil
}

// Html response
func (ctx *Context) Html(code int, html []byte) {
    ctx.Output.Header().Set("Content-Type", "text/html; charset=utf-8")
    ctx.StatusCode(code)
    _, _ = ctx.Output.Write(html)
}

// Json response
func (ctx *Context) Json(code int, m interface{}) {
    ctx.Output.Header().Set("Cache-Control", "no-store")
    ctx.Output.Header().Set("Content-Type", "application/json; charset=utf-8")

    writer := bytes.NewBuffer([]byte{})
    encoder := json.NewEncoder(writer)
    if err := encoder.Encode(&m); err != nil {
        code = http.StatusInternalServerError
    }

    ctx.StatusCode(code)
    _, _ = ctx.Output.Write(writer.Bytes())
}

// CanonicalJson response
func (ctx *Context) CanonicalJson(code int, msg string, data interface{}) {
    ctx.Output.Header().Set("Cache-Control", "no-store")
    ctx.Output.Header().Set("Content-Type", "application/json; charset=utf-8")

    arr := make(Array)
    arr["code"] = code
    arr["msg"] = msg
    arr["data"] = data

    b, err := json.Marshal(&arr)
    if err != nil {
        b = []byte(`{"code":500,"msg":"json encode error","data":""}`)
    }

    ctx.StatusCode(code)
    _, _ = ctx.Output.Write(b)
}

// Text response
func (ctx *Context) Text(code int, text []byte) {
    ctx.Output.Header().Set("Content-Type", "text/plain; charset=utf-8")
    ctx.StatusCode(code)
    _, _ = ctx.Output.Write(text)
}

// StatusCode
func (ctx *Context) StatusCode(code int) {
    ctx.code = code
    ctx.Output.WriteHeader(ctx.code)
}

// Redirect response
func (ctx *Context) Redirect(code int, url string) {
    if 300 < code && code < 400 {
        ctx.StatusCode(code)
        ctx.Output.Header().Set("Location", url)
    }
}

// HttpsUpgrade redirect response
func (ctx *Context) HttpsUpgrade(port int) {
    url := "https://" + ctx.GetHost()
    if port != 443 {
        url += ":" + strconv.FormatInt(int64(port), 10) + ctx.Input.URL.Path
    }
    if len(ctx.Input.URL.RawQuery) > 0 {
        url += ctx.Input.URL.Path + "?" + ctx.Input.URL.RawQuery
    } else {
        url += ctx.Input.URL.Path
    }
    ctx.Redirect(http.StatusMovedPermanently, url)
}

// Cors response
func (ctx *Context) Cors(domain []string) {
    allow := false
    origin := ctx.Input.Header.Get("Origin")
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
        ctx.StatusCode(http.StatusNoContent)
        ctx.Output.Header().Set("Access-Control-Allow-Origin", origin)
        if ctx.Input.Method == http.MethodOptions {
            ctx.Output.Header().Set("Access-Control-Allow-Methods", strings.Join(method, ", "))
            ctx.Output.Header().Set("Access-Control-Allow-Headers", ctx.Input.Header.Get("Access-Control-Request-Headers"))
        }
    }
}

// Error page response
func (ctx *Context) Error(code int) {
    if handler, exist := errorHandler[code]; exist {
        handler(ctx)
    } else {
        ctx.StatusCode(code)
        _, _ = ctx.Output.Write([]byte(http.StatusText(code)))
    }
}

// FileServerHandler returns the file server handler
func FileServerHandler(prefix, dir string) HandlerFunc {
    return func(httpCtx *Context) {
        http.StripPrefix(prefix, http.FileServer(http.Dir(dir))).ServeHTTP(httpCtx.Output, httpCtx.Input)
    }
}

// GetErrorHandler returns the specified error handler
func GetErrorHandler(code int) HandlerFunc {
    if handler, exist := errorHandler[code]; exist {
        return handler
    }

    return func(ctx *Context) {
        ctx.StatusCode(code)
        _, _ = ctx.Output.Write([]byte(http.StatusText(code)))
    }
}

// SetErrorHandler for custom error handler
func SetErrorHandler(code int, handler HandlerFunc) {
    if handler != nil {
        errorHandler[code] = handler
    }
}

// Handler is used to execute the handler
func (ctx *Context) Handler(handler HandlerFunc) {
    handler(ctx)
}

// Logger
func (ctx *Context) Logger(t time.Time) string {
    return fmt.Sprintf("%s %s %s %s %v %d %s", ctx.GetClientIp(), ctx.Input.Method, ctx.Input.Host, ctx.Input.URL, time.Since(t), ctx.code, ctx.Input.UserAgent())
}

// GetCookie
func (ctx *Context) GetCookie(key string) (*http.Cookie, error) {
    return ctx.Input.Cookie(key)
}

// SetCookie
func (ctx *Context) SetCookie(cookie *http.Cookie) {
    http.SetCookie(ctx.Output, cookie)
}

// DelCookie
func (ctx *Context) DelCookie(key string) {
    cookie := &http.Cookie{
        Name:    key,
        Path:    "/",
        MaxAge:  0,
        Expires: time.Now().AddDate(-1, 0, 0),
    }

    http.SetCookie(ctx.Output, cookie)
}
