package router

type Helper struct {
    router *Router
    group  string
    target string
}

// Group
func (r *Router) Group(group string) *Helper {
    length := len(group)
    if length > 0 {
        // make sure to start with /
        if group[0] != '/' {
            group = "/" + group
        }
        // make sure to end with /
        if group[length-1] != '/' {
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

// Use
func (h *Helper) Use(middleware ...HandlerFunc) *Helper {
    h.router.tree.Insert("ANY", h.group, nil, middleware...)

    return h
}

// Target
func (h *Helper) Target(target string) *Helper {
    // make sure not to start with /
    if len(target) > 0 && target[0] == '/' {
        target = target[1:]
    }
    h.target = target

    return h
}

// Head
func (h *Helper) Head(handler HandlerFunc) *Helper {
    h.router.Head(h.group+h.target, handler)

    return h
}

// Options
func (h *Helper) Options(handler HandlerFunc) *Helper {
    h.router.Options(h.group+h.target, handler)

    return h
}

// Get
func (h *Helper) Get(handler HandlerFunc) *Helper {
    h.router.Get(h.group+h.target, handler)

    return h
}

// Post
func (h *Helper) Post(handler HandlerFunc) *Helper {
    h.router.Post(h.group+h.target, handler)

    return h
}

// Put
func (h *Helper) Put(handler HandlerFunc) *Helper {
    h.router.Put(h.group+h.target, handler)

    return h
}

// Delete
func (h *Helper) Delete(handler HandlerFunc) *Helper {
    h.router.Delete(h.group+h.target, handler)

    return h
}

// Any
func (h *Helper) Any(handler HandlerFunc) *Helper {
    h.router.Any(h.group+h.target, handler)

    return h
}
