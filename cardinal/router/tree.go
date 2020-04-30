package router

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
    children map[string]*Node
    handle   map[string]HandlerFunc
}

// NewNode returns a new node based on the name and node depth
func NewNode(name string, depth int) *Node {
    return &Node{
        name:     name,
        depth:    depth,
        children: make(map[string]*Node),
        handle:   make(map[string]HandlerFunc),
    }
}

// Insert a routing rule into the tree
func (t *Tree) Insert(method, pattern string, handle HandlerFunc) {
    // TODO://
}

// Find the controller in the tree according to the routing rule
func (t *Tree) Find(pattern string) *Node {
    // TODO://
    return nil
}
