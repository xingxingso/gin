package gin

import "github.com/xingxingso/gin/internal/bytesconv"

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

func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}

func longestCommonPrefix(a, b string) int {
	i := 0
	max := min(len(a), len(b))
	for i < max && a[i] == b[i] {
		i++
	}
	return i
}

type nodeType uint8

const (
	static nodeType = iota
	root
	param
	catchAll
)

type node struct {
	path    string
	indices string
	//wildChild bool
	nType nodeType
	//priority  uint32
	children []*node // child nodes, at most 1 :param style node at the end of the array
	handlers HandlersChain
	fullPath string
}

// addChild will add a child node, keeping wildcardChild at the end
func (n *node) addChild(child *node) {
	//if n.wildChild && len(n.children) > 0 {
	//	wildcardChild := n.children[len(n.children)-1]
	//	n.children = append(n.children[:len(n.children)-1], child, wildcardChild)
	//} else {
	n.children = append(n.children, child)
	//}
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

	parentFullPathIndex := 0

walk:
	for {
		// Find the longest common prefix.
		// This also implies that the common prefix contains no ':' or '*'
		// since the existing key can't contain those chars.
		i := longestCommonPrefix(path, n.path)

		// Split edge
		if i < len(n.path) {
			child := node{
				path: n.path[i:],
				//wildChild: n.wildChild,
				nType:    static,
				indices:  n.indices,
				children: n.children,
				handlers: n.handlers,
				//priority: n.priority - 1,
				fullPath: n.fullPath,
			}

			n.children = []*node{&child}
			// []byte for proper unicode char conversion, see #65
			n.indices = bytesconv.BytesToString([]byte{n.path[i]})
			n.path = path[:i]
			n.handlers = nil
			//n.wildChild = false
			n.fullPath = fullPath[:parentFullPathIndex+i]
		}

		// Make new node a child of this node
		if i < len(path) {
			path = path[i:]
			c := path[0]

			// Check if a child with the next path byte exists
			for i, max := 0, len(n.indices); i < max; i++ {
				if c == n.indices[i] {
					parentFullPathIndex += len(n.path)
					//i = n.incrementChildPrio(i)
					n = n.children[i]
					continue walk
				}
			}

			// Otherwise insert it
			if c != ':' && c != '*' && n.nType != catchAll {
				// []byte for proper unicode char conversion, see #65
				n.indices += bytesconv.BytesToString([]byte{c})
				child := &node{
					fullPath: fullPath,
				}
				n.addChild(child)
				//n.incrementChildPrio(len(n.indices) - 1)

				n = child
			} /*else if n.wildChild {
				// inserting a wildcard node, need to check if it conflicts with the existing wildcard
				n = n.children[len(n.children)-1]
				n.priority++

				// Check if the wildcard matches
				if len(path) >= len(n.path) && n.path == path[:len(n.path)] &&
					// Adding a child to a catchAll is not possible
					n.nType != catchAll &&
					// Check for longer wildcard, e.g. :name and :names
					(len(n.path) >= len(path) || path[len(n.path)] == '/') {
					continue walk
				}

				// Wildcard conflict
				pathSeg := path
				if n.nType != catchAll {
					pathSeg = strings.SplitN(pathSeg, "/", 2)[0]
				}
				prefix := fullPath[:strings.Index(fullPath, pathSeg)] + n.path
				panic("'" + pathSeg +
					"' in new path '" + fullPath +
					"' conflicts with existing wildcard '" + n.path +
					"' in existing prefix '" + prefix +
					"'")
			}*/

			n.insertChild(path, fullPath, handlers)
			return
		}

		// Otherwise add handle to current node
		if n.handlers != nil {
			panic("handlers are already registered for path '" + fullPath + "'")
		}
		n.handlers = handlers
		n.fullPath = fullPath
		return
	}
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

walk: // Outer loop for walking the tree
	for {
		prefix := n.path

		if len(path) > len(prefix) {
			if path[:len(prefix)] == prefix {
				path = path[len(prefix):]

				//// Try all the non-wildcard children first by matching the indices
				idxc := path[0]
				for i, c := range []byte(n.indices) {
					if c == idxc {
						n = n.children[i]
						continue walk
					}
				}

				//if !n.wildChild {
				// If the path at the end of the loop is not equal to '/' and the current node has no child nodes
				// the current node needs to roll back to last valid skippedNode
				//if path != "/" {
				//	for length := len(*skippedNodes); length > 0; length-- {
				//		skippedNode := (*skippedNodes)[length-1]
				//		*skippedNodes = (*skippedNodes)[:length-1]
				//		if strings.HasSuffix(skippedNode.path, path) {
				//			path = skippedNode.path
				//			n = skippedNode.node
				//			if value.params != nil {
				//				*value.params = (*value.params)[:skippedNode.paramsCount]
				//			}
				//			globalParamsCount = skippedNode.paramsCount
				//			continue walk
				//		}
				//	}
				//}

				// Nothing found.
				// We can recommend to redirect to the same URL without a
				// trailing slash if a leaf exists for that path.
				value.tsr = path == "/" && n.handlers != nil
				return
				//}

				// Handle wildcard child, which is always at the end of the array
				//n = n.children[len(n.children)-1]
				//globalParamsCount++

				//switch n.nType {
				//case param:
				//	// fix truncate the parameter
				//	// tree_test.go  line: 204
				//
				//	// Find param end (either '/' or path end)
				//	end := 0
				//	for end < len(path) && path[end] != '/' {
				//		end++
				//	}
				//
				//	// Save param value
				//	if params != nil && cap(*params) > 0 {
				//		if value.params == nil {
				//			value.params = params
				//		}
				//		// Expand slice within preallocated capacity
				//		i := len(*value.params)
				//		*value.params = (*value.params)[:i+1]
				//		val := path[:end]
				//		if unescape {
				//			if v, err := url.QueryUnescape(val); err == nil {
				//				val = v
				//			}
				//		}
				//		(*value.params)[i] = Param{
				//			Key:   n.path[1:],
				//			Value: val,
				//		}
				//	}
				//
				//	// we need to go deeper!
				//	if end < len(path) {
				//		if len(n.children) > 0 {
				//			path = path[end:]
				//			n = n.children[0]
				//			continue walk
				//		}
				//
				//		// ... but we can't
				//		value.tsr = len(path) == end+1
				//		return
				//	}
				//
				//	if value.handlers = n.handlers; value.handlers != nil {
				//		value.fullPath = n.fullPath
				//		return
				//	}
				//	if len(n.children) == 1 {
				//		// No handle found. Check if a handle for this path + a
				//		// trailing slash exists for TSR recommendation
				//		n = n.children[0]
				//		value.tsr = (n.path == "/" && n.handlers != nil) || (n.path == "" && n.indices == "/")
				//	}
				//	return
				//
				//case catchAll:
				//	// Save param value
				//	if params != nil {
				//		if value.params == nil {
				//			value.params = params
				//		}
				//		// Expand slice within preallocated capacity
				//		i := len(*value.params)
				//		*value.params = (*value.params)[:i+1]
				//		val := path
				//		if unescape {
				//			if v, err := url.QueryUnescape(path); err == nil {
				//				val = v
				//			}
				//		}
				//		(*value.params)[i] = Param{
				//			Key:   n.path[2:],
				//			Value: val,
				//		}
				//	}
				//
				//	value.handlers = n.handlers
				//	value.fullPath = n.fullPath
				//	return
				//
				//default:
				//	panic("invalid node type")
				//}
			}
		}

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
			//for i, c := range []byte(n.indices) {
			//	if c == '/' {
			//		n = n.children[i]
			//		value.tsr = (len(n.path) == 1 && n.handlers != nil) ||
			//			(n.nType == catchAll && n.children[0].handlers != nil)
			//		return
			//	}
			//}

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
