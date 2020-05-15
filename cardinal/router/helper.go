package router

import (
    "net/http"
    "strings"
)

type Helper struct {
    router *Router
    group  string
    target string
}

func (r *Router) Group(group string) *Helper {
    if len(group) > 0 {
        // make sure to start with /
        if !strings.HasPrefix(group, "/") {
            group = "/" + group
        }
        // make sure to end with /
        if !strings.HasSuffix(group, "/") {
            group = group + "/"
        }
    } else {
        group = "/"
    }

    return &Helper{
        router: r,
        group:  group,
    }
}

func (h *Helper) Use(middleware ...MiddlewareFunc) *Helper {
    h.router.tree.Insert("ANY", h.group, nil, middleware...)

    return h
}

func (h *Helper) Target(target string) *Helper {
    // make sure not to start with /
    if strings.HasPrefix(target, "/") {
        target = target[1:]
    }
    h.target = target

    return h
}

func (h *Helper) Head(handler HandlerFunc) *Helper {
    h.router.Head(h.group+h.target, handler)

    return h
}

func (h *Helper) Options(handler HandlerFunc) *Helper {
    h.router.Options(h.group+h.target, handler)

    return h
}

func (h *Helper) Get(handler HandlerFunc) *Helper {
    h.router.Get(h.group+h.target, handler)

    return h
}

func (h *Helper) Post(handler HandlerFunc) *Helper {
    h.router.Post(h.group+h.target, handler)

    return h
}

func (h *Helper) Put(handler HandlerFunc) *Helper {
    h.router.Put(h.group+h.target, handler)

    return h
}

func (h *Helper) Delete(handler HandlerFunc) *Helper {
    h.router.Delete(h.group+h.target, handler)

    return h
}

func (h *Helper) Any(handler HandlerFunc) *Helper {
    h.router.Any(h.group+h.target, handler)

    return h
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
