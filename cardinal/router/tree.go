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
    name     string
    depth    int
    fullName string
    children map[string]*Node
    handle   map[string]HandlerFunc
}

// NewNode returns a new node based on the name and node depth
func NewNode(name string, depth int) *Node {
    fullName := ""
    if name == "/" {
        fullName = "/"
    }

    return &Node{
        name:     name,
        depth:    depth,
        fullName: fullName,
        children: make(map[string]*Node),
        handle:   make(map[string]HandlerFunc),
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

func getPattern(rule string) (key, pattern string) {
    length := len(rule)
    firstChar := rule[:1]
    lastChar := rule[length-1:]

    if firstChar == ":" {
        // e.g. :id, :name, :key
        key = rule[1:]
        if key == "id" {
            pattern = `^[\d]+$`
        } else if key == "name" {
            pattern = `^[\w]+$`
        }
    } else if length > 2 && firstChar == "{" && lastChar == "}" {
        // e.g. {id:^[\d]+$}, {name}, {*}
        res := strings.Split(rule[1:length-1], ":")
        key = res[0]
        if res[0] != "*" && len(res) > 1 && res[1] != "" {
            pattern = res[1]
        }
    }

    return
}
