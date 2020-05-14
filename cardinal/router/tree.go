package router

import (
    "regexp"
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

// Node mounts the callback controller
type Node struct {
    depth    int
    rule     string
    fullRule string
    key      string
    pattern  *regexp.Regexp
    parent   *Node
    children map[string]*Node
    handlers map[string]HandlerFunc
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
func (t *Tree) Insert(method, pattern string, handle HandlerFunc) {
    currentNode := t.root
    currentFullName := t.root.fullName

    if pattern != currentNode.name {
        names := strings.Split(pattern, "/")
        length := len(names)
        for index, name := range names {
            if name == "" {
                // exclude empty name that begin or end with /
                continue
            }

            currentFullName += name
            // according to the index to determine whether it is a path name
            if index < length-1 {
                currentFullName += "/"
            }

            node, exist := currentNode.children[name]
            if !exist {
                node = NewNode(name, currentNode.depth+1)
                node.fullName = currentFullName
                currentNode.children[name] = node
            }

            currentNode = node

            // do not register nodes after {*} nodes
            if name == "{*}" {
                break
            }
        }
    }

    // register the controller method at the last node
    currentNode.handle[strings.ToUpper(method)] = handle
}

// Query the controller in the tree according to the request uri
func (t *Tree) Query(requestUri string) (*Node, map[string]string) {
    currentNode := t.root
    currentRequestUri := t.root.fullName
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

        currentRequestUri += name
        if index < length-1 {
            currentRequestUri += "/"
        }

        node, exist := currentNode.children[name]
        if exist {
            currentNode = node
        } else {
            // find and match rules
            found := false
            for rule, childNode := range currentNode.children {
                key, pattern := getPattern(rule)
                if key != "" {
                    if key == "*" {
                        params[key] = requestUri[strings.Index(requestUri, name):]

                        return childNode, params
                    } else {
                        // may be panic here if the regexp is wrong
                        if pattern != "" && !regexp.MustCompile(pattern).MatchString(name) {
                            // Irregular and continue to the next match
                            continue
                        }

                        params[key] = name
                        currentNode = childNode
                        found = true

                        // meet the rules and jump out of the current node's match
                        break
                    }
                }
            }

            if !found {
                return nil, params
            }
        }
    }

    return currentNode, params
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
