package router

import (
    "fmt"
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

// Node mounts the callback controller
type Node struct {
    depth      int
    rule       string
    fullRule   string
    key        string
    pattern    *regexp.Regexp
    parent     *Node
    children   map[string]*Node
    handlers   map[string]HandlerFunc
    middleware []HandlerFunc
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
        handlers: make(map[string]HandlerFunc),
    }
}

// Insert a routing rule into the tree
func (t *Tree) Insert(method, fullRule string, handler HandlerFunc, middleware ...HandlerFunc) {
    currentNode := t.root
    currentFullRule := t.root.fullRule

    if currentFullRule != fullRule {
        length := len(fullRule)
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
                node.pattern = pattern
                node.parent = currentNode
                currentNode.children[rule] = node
            }

            currentNode = node

            // do not register nodes after wildcard nodes
            if isWildcard {
                break
            }

            start = i + 1
        }
    }

    // register the controller method at the last node
    if handler != nil {
        currentNode.handlers[strings.ToUpper(method)] = handler
    }
    for _, m := range middleware {
        if m != nil {
            currentNode.middleware = append(currentNode.middleware, m)
        }
    }
}

// Match the request uri in the tree to get the target node
func (t *Tree) Match(requestUri string) (*Node, map[string]string) {
    currentNode := t.root
    currentRequestUri := t.root.fullRule
    params := make(map[string]string)

    if currentRequestUri == requestUri {
        return currentNode, params
    }

    names := strings.Split(requestUri, "/")
    length := len(names)
    for index, name := range names {
        if name == "" {
            continue
        }

        node, exist := currentNode.children[name]
        if !exist {
            // match the rule
            found := false
            for _, childNode := range currentNode.children {
                if childNode.key[len(childNode.key)-1:] == "*" {
                    if childNode.key != "*" && !strings.HasPrefix(name, childNode.key[:len(childNode.key)-1]) {
                        continue
                    }

                    // for wildcard, e.g. *, static*
                    params[childNode.key] = requestUri[len(currentRequestUri):]

                    return childNode, params
                } else if childNode.key != childNode.rule {
                    // for custom rule
                    if childNode.pattern != nil && !childNode.pattern.MatchString(name) {
                        // rule mismatch and continue to the next match
                        continue
                    }

                    // the rule has been successfully matched
                    found = true
                    params[childNode.key] = name
                    node = childNode

                    // jumps out of matching child nodes at the current node
                    break
                }
            }

            if !found {
                return nil, params
            }
        }

        currentRequestUri += name
        if index < length-1 {
            currentRequestUri += "/"
        }

        currentNode = node
    }

    return currentNode, params
}

// PrintRoutes print the controller and middleware for each routing rule
func (t *Tree) PrintRoutes(node *Node) {
    if node == nil {
        node = t.root
    }

    fn := func(node *Node) {
        p := node.fullRule
        if len(node.handlers) > 0 {
            // middleware list
            middleware := MiddlewareList(node)
            if len(middleware) > 0 {
                p += " ["
                for i, m := range middleware {
                    p += " " + strconv.Itoa(i) + ":" + runtime.FuncForPC(reflect.ValueOf(m).Pointer()).Name()
                }
                p += " ]"
            }

            // controller list
            p += " ["
            for method, fn := range node.handlers {
                p += " " + method + ":" + runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name()
            }
            p += " ]"

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
        // e.g. :id, :name, :uid(^[\d]+$), :*, :static*, :str
        s := rule[1:]
        if s == "id" {
            key = "id"
            pattern = regexp.MustCompile(`^[\d]+$`)
        } else if s == "name" {
            key = "name"
            pattern = regexp.MustCompile(`^[\w-]+$`)
        } else {
            a := strings.Index(s, "(")
            b := strings.LastIndex(s, ")")
            if s[len(s)-1:] == "*" {
                // *, static*
                isWildcard = true
                key = s
            } else if 0 < a && a < b {
                // :uid(^[\d]+$)
                key = s[:a]
                pattern = regexp.MustCompile(s[a+1 : b])
            } else if a == b {
                // :str
                key = s
            }
        }
    } else if firstChar == "{" && lastChar == "}" && length > 2 {
        // e.g. {id}, {name}, {uid:^[\d]+$}, {str}, {*}, {static*}
        res := strings.Split(rule[1:length-1], ":")
        key = res[0]
        if key == "id" {
            pattern = regexp.MustCompile(`^[\d]+$`)
        } else if key == "name" {
            pattern = regexp.MustCompile(`^[\w-]+$`)
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

// MiddlewareList returns middleware in each layer in top-down order
func MiddlewareList(node *Node) []HandlerFunc {
    if node.fullRule == "/" {
        return node.middleware
    }

    list := make([]HandlerFunc, 0)
    var fn func(node *Node) []HandlerFunc
    fn = func(node *Node) []HandlerFunc {
        if node.parent != nil {
            if len(node.parent.middleware) > 0 {
                list = append(node.parent.middleware, list...)
            }

            node = node.parent
            fn(node)
        }

        return list
    }

    return fn(node)
}
