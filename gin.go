package gin

import (
	"net/http"
	"sync"
)

var (
	default404Body = []byte("404 page not found")
	default405Body = []byte("405 method not allowed")
)

// HandlerFunc defines the handler used by gin middleware as return value.
type HandlerFunc func(*Context)

// HandlersChain defines a HandlerFunc slice.
type HandlersChain []HandlerFunc

// Engine is the framework's instance, it contains the muxer, middleware and configuration settings.
// Create an instance of Engine, by using New() or Default()
type Engine struct {
	RouterGroup

	pool      sync.Pool
	maxParams uint16
}

// New returns a new blank Engine instance without any middleware attached.
// By default, the configuration is:
// - RedirectTrailingSlash:  true
// - RedirectFixedPath:      false
// - HandleMethodNotAllowed: false
// - ForwardedByClientIP:    true
// - UseRawPath:             false
// - UnescapePathValues:     true
func New() *Engine { // todo
	engine := &Engine{}
	engine.pool.New = func() any {
		return engine.allocateContext(engine.maxParams)
	}
	return engine
}

// Default returns an Engine instance with the Logger and Recovery middleware already attached.
func Default() *Engine {
	engine := New()
	return engine
}

func (engine *Engine) Handler() http.Handler {
	return engine
}

func (engine *Engine) allocateContext(maxParams uint16) *Context {
	return &Context{}
}

// Run attaches the router to a http.Server and starts listening and serving HTTP requests.
// It is a shortcut for http.ListenAndServe(addr, router)
// Note: this method will block the calling goroutine indefinitely unless an error happens.
func (engine *Engine) Run(addr ...string) (err error) {
	address := resolveAddress(addr)
	err = http.ListenAndServe(address, engine.Handler())
	return
}

// ServeHTTP conforms to the http.Handler interface.
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := engine.pool.Get().(*Context)
	c.writermem.reset(w)
	c.Request = req
	c.reset()

	engine.handleHTTPRequest(c)

	engine.pool.Put(c)
}

func (engine *Engine) handleHTTPRequest(c *Context) {
	//httpMethod := c.Request.Method
	//rPath := c.Request.URL.Path
	//unescape := false
	//if engine.UseRawPath && len(c.Request.URL.RawPath) > 0 {
	//	rPath = c.Request.URL.RawPath
	//	unescape = engine.UnescapePathValues
	//}

	//if engine.RemoveExtraSlash {
	//	rPath = cleanPath(rPath)
	//}

	// Find root of the tree for the given HTTP method
	//t := engine.trees
	//for i, tl := 0, len(t); i < tl; i++ {
	//	if t[i].method != httpMethod {
	//		continue
	//	}
	//	root := t[i].root
	//	// Find route in tree
	//	value := root.getValue(rPath, c.params, c.skippedNodes, unescape)
	//	if value.params != nil {
	//		c.Params = *value.params
	//	}
	//	if value.handlers != nil {
	//		c.handlers = value.handlers
	//		c.fullPath = value.fullPath
	//		c.Next()
	//		c.writermem.WriteHeaderNow()
	//		return
	//	}
	//	if httpMethod != http.MethodConnect && rPath != "/" {
	//		if value.tsr && engine.RedirectTrailingSlash {
	//			redirectTrailingSlash(c)
	//			return
	//		}
	//		if engine.RedirectFixedPath && redirectFixedPath(c, root, engine.RedirectFixedPath) {
	//			return
	//		}
	//	}
	//	break
	//}

	//if engine.HandleMethodNotAllowed {
	//	for _, tree := range engine.trees {
	//		if tree.method == httpMethod {
	//			continue
	//		}
	//		if value := tree.root.getValue(rPath, nil, c.skippedNodes, unescape); value.handlers != nil {
	//			c.handlers = engine.allNoMethod
	//			serveError(c, http.StatusMethodNotAllowed, default405Body)
	//			return
	//		}
	//	}
	//}
	//c.handlers = engine.allNoRoute
	serveError(c, http.StatusNotFound, default404Body)
}

var mimePlain = []string{MIMEPlain}

func serveError(c *Context, code int, defaultMessage []byte) {
	c.writermem.status = code
	//c.Next()
	if c.writermem.Written() {
		return
	}
	if c.writermem.Status() == code {
		c.writermem.Header()["Content-Type"] = mimePlain
		_, err := c.Writer.Write(defaultMessage)
		if err != nil {
			debugPrint("cannot write message to writer during serve error: %v", err)
		}
		return
	}
	c.writermem.WriteHeaderNow()
}
