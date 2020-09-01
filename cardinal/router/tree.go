package router

import (
    "fmt"
    "net/http"
    "reflect"
    "regexp"
    "runtime"
    "strconv"
    "strings"
)

// Tree is a prefix tree that routing rules with the same namespace
// will share the same prefix node, a bit like Trie
type Tree struct {
    root *Node
}

// NewTree returns a new prefix tree
func NewTree() *Tree {
    return &Tree{
        // create root node
        root: NewNode("/", 0),
    }
}

// HandlerFunc is the type of middleware and controller
type HandlerFunc func(httpCtx *Context)

type HandlersChain []HandlerFunc

// Node mounts the callback controller
type Node struct {
    depth      int
    rule       string
    fullRule   string
    key        string
    isWildcard bool
    pattern    *regexp.Regexp
    parent     *Node
    children   map[string]*Node
    handlers   map[string]HandlersChain
    middleware HandlersChain
}

// NewNode returns a new node based on the rule and node depth
func NewNode(rule string, depth int) *Node {
    fullRule := ""
    if rule == "/" {
        fullRule = "/"
    }

    return &Node{
        depth:    depth,
        rule:     rule,
        fullRule: fullRule,
        children: make(map[string]*Node),
        handlers: make(map[string]HandlersChain),
    }
}

// Insert a routing rule into the tree
func (t *Tree) Insert(method, fullRule string, handler HandlerFunc, middleware ...HandlerFunc) {
    currentNode := t.root
    // always start with /
    if fullRule == "" || fullRule[0] != '/' {
        fullRule = "/" + fullRule
    }
    length := len(fullRule)

    var handlers HandlersChain
    if currentNode.fullRule != fullRule {
        start := 1
        for i := start; i <= length; i++ {
            if i < length && fullRule[i] != '/' {
                continue
            }
            rule := fullRule[start:i]
            if rule == "" {
                continue
            }

            // check and parse the rule
            // may be panic here if the rule is wrong
            key, pattern, isWildcard := compile(rule)
            if key == "" {
                panic("router: rule compile error: `" + rule + "` in " + fullRule)
            }

            node, exist := currentNode.children[rule]
            if !exist {
                node = NewNode(rule, currentNode.depth+1)
                node.fullRule = fullRule[:i]
                node.key = key
                node.isWildcard = isWildcard
                node.pattern = pattern
                node.parent = currentNode
                currentNode.children[rule] = node
            }

            currentNode = node

            // save the passed middleware to the handlers
            // handlers will be used directly when the route matching hits
            if handler != nil && len(currentNode.middleware) > 0 {
                handlers = append(handlers, currentNode.middleware...)
            }

            // do not register nodes after wildcard nodes
            if isWildcard {
                break
            }

            start = i + 1
        }
    }

    // filter
    for _, m := range middleware {
        if m != nil {
            currentNode.middleware = append(currentNode.middleware, m)
        }
    }

    // register the controller method at the last node
    if handler != nil {
        handlers = append(handlers, currentNode.middleware...)
        handlers = append(handlers, handler)
        currentNode.handlers[strings.ToUpper(method)] = handlers
    }
}

// Match the request uri in the tree to get the target node
func (t *Tree) Match(ctx *Context, requestUri, method string) {
    currentNode := t.root
    ctx.params = make(map[string]string)
    if currentNode.fullRule != requestUri {
        length := len(requestUri)
        start := 1
        for i := start; i <= length; i++ {
            if i < length && requestUri[i] != '/' {
                continue
            }
            name := requestUri[start:i]
            if name == "" {
                continue
            }

            // match a node that meets the rules in children nodes
            node, found := currentNode.children[name]
            if !found {
                for _, childNode := range currentNode.children {
                    if childNode.key != childNode.rule {
                        if childNode.isWildcard {
                            // for wildcard
                            if childNode.key == "*" || childNode.key == name+"*" {
                                found = true
                                ctx.params[childNode.key] = requestUri[start:]
                                node = childNode

                                break
                            }
                        } else {
                            // for regexp rule
                            if childNode.pattern != nil {
                                result := childNode.pattern.FindStringSubmatch(name)
                                if len(result) == 2 {
                                    found = true
                                    ctx.params[childNode.key] = result[1]
                                    node = childNode

                                    break
                                }
                            }
                        }
                    }
                }

                // node not found
                if !found {
                    currentNode = nil

                    break
                }
            }

            // node found
            currentNode = node
            if currentNode.isWildcard {
                // do not match nodes after wildcard nodes
                break
            }

            start = i + 1
        }
    }

    if currentNode != nil {
        var exist bool
        ctx.handlers, exist = currentNode.handlers[method]
        if ctx.handlers == nil {
            // call the GET handler if the HEAD handler does not exist
            if method == http.MethodHead {
                ctx.handlers, exist = currentNode.handlers[http.MethodGet]
            }

            if ctx.handlers == nil {
                // default handler
                ctx.handlers, exist = currentNode.handlers["ANY"]
                if ctx.handlers == nil {
                    ctx.handlers = append(ctx.handlers, t.root.middleware...)
                    if exist {
                        ctx.handlers = append(ctx.handlers, GetErrorHandler(http.StatusNotImplemented))
                    } else {
                        ctx.handlers = append(ctx.handlers, GetErrorHandler(http.StatusMethodNotAllowed))
                    }
                }
            }
        }
    } else {
        // not found
        ctx.handlers = append(ctx.handlers, t.root.middleware...)
        ctx.handlers = append(ctx.handlers, GetErrorHandler(http.StatusNotFound))
    }

    return
}

// PrintRoutes print the controller and middleware for each routing rule
func (t *Tree) PrintRoutes(node *Node) {
    if node == nil {
        node = t.root
    }

    fn := func(node *Node) {
        var p string
        if len(node.handlers) > 0 {
            for method, chain := range node.handlers {
                p += method + " " + node.fullRule
                p += " ["
                for i, fn := range chain {
                    name := runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
                    if len(chain)-1 != i || strings.Contains(name, ".func") {
                        name = name[:len(name)-6]
                        index := strings.LastIndex(name, ".")
                        name = name[index+1:]
                    }
                    p += " " + strconv.Itoa(i) + ":" + name
                }
                p += " ]"
            }

            fmt.Println(p)
        }
    }

    fn(node)
    if len(node.children) > 0 {
        for _, child := range node.children {
            t.PrintRoutes(child)
        }
    }
}

// parse the rule and compile it
func compile(rule string) (key string, pattern *regexp.Regexp, isWildcard bool) {
    length := len(rule)
    firstChar := rule[:1]
    lastChar := rule[length-1:]

    if firstChar == ":" && length > 1 {
        // e.g. :id, :name, :uid(^([\d]+)$), :*, :static*, :str
        s := rule[1:]
        if s == "id" {
            key = "id"
            pattern = regexp.MustCompile(`^([\d]+)$`)
        } else if s == "name" {
            key = "name"
            pattern = regexp.MustCompile(`^([\w-]+)$`)
        } else if s[len(s)-1:] == "*" {
            // :*, :static*
            isWildcard = true
            key = s
        } else {
            a := strings.Index(s, "(")
            b := strings.LastIndex(s, ")")
            if 0 < a && a < b {
                // :uid(^([\d]+)$)
                key = s[:a]
                pattern = regexp.MustCompile(s[a+1 : b])
            } else if a == -1 && b == -1 {
                // :str
                key = s
            }
        }
    } else if firstChar == "{" && lastChar == "}" && length > 2 {
        // e.g. {id}, {name}, {uid:^[\d]+$}, {str}, {*}, {static*}
        res := strings.Split(rule[1:length-1], ":")
        key = res[0]
        if key == "id" {
            pattern = regexp.MustCompile(`^([\d]+)$`)
        } else if key == "name" {
            pattern = regexp.MustCompile(`^([\w-]+)$`)
        } else if key[len(key)-1:] == "*" {
            // {*}, {static*}
            isWildcard = true
        } else if len(res) > 1 && res[1] != "" {
            pattern = regexp.MustCompile(res[1])
        }
    } else if length > 0 && firstChar != ":" && firstChar != "{" {
        key = rule
    }

    return
}
