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

// HandlerFunc is the type of controller
type HandlerFunc func(httpCtx *Context)

// MiddlewareFunc is the type of middleware
type MiddlewareFunc func(handler HandlerFunc) HandlerFunc

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
    middleware []MiddlewareFunc
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
func (t *Tree) Insert(method, fullRule string, handler HandlerFunc, middleware ...MiddlewareFunc) {
    currentNode := t.root
    currentFullRule := t.root.fullRule

    if currentFullRule != fullRule {
        rules := strings.Split(fullRule, "/")
        length := len(rules)
        for index, rule := range rules {
            if rule == "" {
                // exclude empty rule that begin or end with / and consecutive /
                continue
            }

            // check and parse the rule
            // may be panic here if the rule is wrong
            key, pattern := compile(rule)
            if key == "" {
                panic("rule error: `" + rule + "` in " + fullRule)
            }

            currentFullRule += rule
            // according to the index to determine whether it is a path name
            if index < length-1 {
                currentFullRule += "/"
            }

            node, exist := currentNode.children[rule]
            if !exist {
                node = NewNode(rule, currentNode.depth+1)
                node.fullRule = currentFullRule
                node.key = key
                node.pattern = pattern
                node.parent = currentNode
                currentNode.children[rule] = node
            }

            currentNode = node

            // do not register nodes after * nodes
            if rule == "{*}" || rule == ":*" {
                break
            }
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
                if childNode.key == "*" {
                    // for wildcard
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

// Show the controller and middleware for each routing rule
func (t *Tree) Show(node *Node) {
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

    if len(node.children) > 0 {
        for _, child := range node.children {
            fn(child)
            t.Show(child)
        }
    }
}

// parse the rule and compile it
func compile(rule string) (key string, pattern *regexp.Regexp) {
    length := len(rule)
    firstChar := rule[:1]
    lastChar := rule[length-1:]

    if firstChar == ":" && length > 1 {
        // e.g. :id, :name, :str, :*, :uid(^[\d]+$)
        s := rule[1:]
        if s == "id" {
            key = "id"
            pattern = regexp.MustCompile(`^[\d]+$`)
        } else if s == "name" {
            key = "name"
            pattern = regexp.MustCompile(`^[\w]+$`)
        } else {
            a := strings.Index(s, "(")
            b := strings.LastIndex(s, ")")
            if s[:1] != "*" && 0 < a && a < b {
                key = s[:a]
                pattern = regexp.MustCompile(s[a+1 : b])
            } else {
                key = s
            }
        }
    } else if firstChar == "{" && lastChar == "}" && length > 2 {
        // e.g. {id:^[\d]+$}, {str}, {*}
        res := strings.Split(rule[1:length-1], ":")
        key = res[0]
        if res[0] != "*" && len(res) > 1 && res[1] != "" {
            pattern = regexp.MustCompile(res[1])
        }
    } else if length > 0 && firstChar != ":" && firstChar != "{" {
        key = rule
    }

    return
}

// MiddlewareList returns middleware in each layer in top-down order
func MiddlewareList(node *Node) []MiddlewareFunc {
    list := make([]MiddlewareFunc, 0)
    var fn func(node *Node) []MiddlewareFunc
    fn = func(node *Node) []MiddlewareFunc {
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
