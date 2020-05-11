package router

import "strings"

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
    uri      string
    children map[string]*Node
    handle   map[string]HandlerFunc
}

// NewNode returns a new node based on the name and node depth
func NewNode(name string, depth int) *Node {
    uri := ""
    if name == "/" {
        uri = "/"
    }

    return &Node{
        name:     name,
        depth:    depth,
        uri:      uri,
        children: make(map[string]*Node),
        handle:   make(map[string]HandlerFunc),
    }
}

// Insert a routing rule into the tree
func (t *Tree) Insert(method, pattern string, handle HandlerFunc) {
    currentNode := t.root
    currentUri := t.root.uri

    if pattern != currentNode.name {
        names := strings.Split(pattern, "/")
        length := len(names)
        for index, name := range names {
            if name == "" {
                // exclude empty name that begin or end with /
                continue
            }

            currentUri += name
            // according to the index to determine whether it is a path name
            if index < length-1 {
                currentUri += "/"
            }

            node, exist := currentNode.children[name]
            if !exist {
                node = NewNode(name, currentNode.depth+1)
                node.uri = currentUri
                currentNode.children[name] = node
            }

            currentNode = node
        }
    }

    // register the controller method at the last node
    currentNode.handle[strings.ToUpper(method)] = handle
}

// Find the controller in the tree according to the routing rule
func (t *Tree) Find(pattern string) *Node {
    // TODO://
    return nil
}
