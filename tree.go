package gin

// Param is a single URL parameter, consisting of a key and a value.
type Param struct {
	Key   string
	Value string
}

// Params is a Param-slice, as returned by the router.
// The slice is ordered, the first URL parameter is also the first slice value.
// It is therefore safe to read values by the index.
type Params []Param

type methodTree struct {
	method string
	root   *node
}

type methodTrees []methodTree

func (trees methodTrees) get(method string) *node {
	for _, tree := range trees {
		if tree.method == method {
			return tree.root
		}
	}
	return nil
}

type nodeType uint8

const (
	static nodeType = iota
	root
	param
	catchAll
)

type node struct {
	path      string
	indices   string
	wildChild bool
	nType     nodeType
	//priority  uint32
	children []*node // child nodes, at most 1 :param style node at the end of the array
	handlers HandlersChain
	fullPath string
}

// addRoute adds a node with the given handle to the path.
// Not concurrency-safe!
func (n *node) addRoute(path string, handlers HandlersChain) {
	fullPath := path

	// Empty tree
	if len(n.path) == 0 && len(n.children) == 0 {
		n.insertChild(path, fullPath, handlers)
		n.nType = root
		return
	}
	return
}

func (n *node) insertChild(path string, fullPath string, handlers HandlersChain) {
	// If no wildcard was found, simply insert the path and handle
	n.path = path
	n.handlers = handlers
	n.fullPath = fullPath
}

// nodeValue holds return values of (*Node).getValue method
type nodeValue struct {
	handlers HandlersChain
	params   *Params
	tsr      bool
	fullPath string
}

//type skippedNode struct {
//	path        string
//	node        *node
//	paramsCount int16
//}

// Returns the handle registered with the given path (key). The values of
// wildcards are saved to a map.
// If no handle can be found, a TSR (trailing slash redirect) recommendation is
// made if a handle exists with an extra (without the) trailing slash for the
// given path.
func (n *node) getValue(path string, params *Params) (value nodeValue) {
	//var globalParamsCount int16

	//walk: // Outer loop for walking the tree
	for {
		prefix := n.path

		if path == prefix {
			// If the current path does not equal '/' and the node does not have a registered handle and the most recently matched node has a child node
			// the current node needs to roll back to last valid skippedNode
			if n.handlers == nil && path != "/" {
				//	n = latestNode.children[len(latestNode.children)-1]
			}
			// We should have reached the node containing the handle.
			// Check if this node has a handle registered.
			if value.handlers = n.handlers; value.handlers != nil {
				value.fullPath = n.fullPath
				return
			}

			if path == "/" && n.nType == static {
				value.tsr = true
				return
			}

			// No handle found. Check if a handle for this path + a
			// trailing slash exists for trailing slash recommendation
			for i, c := range []byte(n.indices) {
				if c == '/' {
					n = n.children[i]
					value.tsr = (len(n.path) == 1 && n.handlers != nil) ||
						(n.nType == catchAll && n.children[0].handlers != nil)
					return
				}
			}

			return
		}

		// Nothing found. We can recommend to redirect to the same URL with an
		// extra trailing slash if a leaf exists for that path
		value.tsr = path == "/" ||
			(len(prefix) == len(path)+1 && prefix[len(path)] == '/' &&
				path == prefix[:len(prefix)-1] && n.handlers != nil)

		return
	}
}
