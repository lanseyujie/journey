package router

import (
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

func (h *Helper) Use(middleware ...HandlerFunc) *Helper {
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
