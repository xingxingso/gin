package gin

import (
	"net/http"
	"regexp"
)

var (
	// regEnLetter matches english letters for http method name
	regEnLetter = regexp.MustCompile("^[A-Z]+$")
)

// IRouter defines all router handle interface includes single and group router.
type IRouter interface {
	IRoutes
	//Group(string, ...HandlerFunc) *RouterGroup
}

// IRoutes defines all router handle interface.
type IRoutes interface {
	//Use(...HandlerFunc) IRoutes

	Handle(string, string, ...HandlerFunc) IRoutes
	//Any(string, ...HandlerFunc) IRoutes
	//GET(string, ...HandlerFunc) IRoutes
	//POST(string, ...HandlerFunc) IRoutes
	//DELETE(string, ...HandlerFunc) IRoutes
	//PATCH(string, ...HandlerFunc) IRoutes
	//PUT(string, ...HandlerFunc) IRoutes
	//OPTIONS(string, ...HandlerFunc) IRoutes
	//HEAD(string, ...HandlerFunc) IRoutes
	//Match([]string, string, ...HandlerFunc) IRoutes
	//
	//StaticFile(string, string) IRoutes
	//StaticFileFS(string, string, http.FileSystem) IRoutes
	//Static(string, string) IRoutes
	//StaticFS(string, http.FileSystem) IRoutes
}

// RouterGroup is used internally to configure router, a RouterGroup is associated with
// a prefix and an array of handlers (middleware).
type RouterGroup struct {
	//Handlers HandlersChain
	//basePath string
	engine *Engine
	//root     bool
}

var _ IRouter = (*RouterGroup)(nil)

func (group *RouterGroup) handle(httpMethod, relativePath string, handlers HandlersChain) IRoutes {
	//absolutePath := group.calculateAbsolutePath(relativePath)
	//handlers = group.combineHandlers(handlers)
	absolutePath := relativePath //todo support relative path
	group.engine.addRoute(httpMethod, absolutePath, handlers)
	return group.returnObj()
}

// Handle registers a new request handle and middleware with the given path and method.
// The last handler should be the real handler, the other ones should be middleware that can and should be shared among different routes.
// See the example code in GitHub.
//
// For GET, POST, PUT, PATCH and DELETE requests the respective shortcut
// functions can be used.
//
// This function is intended for bulk loading and to allow the usage of less
// frequently used, non-standardized or custom methods (e.g. for internal
// communication with a proxy).
func (group *RouterGroup) Handle(httpMethod, relativePath string, handlers ...HandlerFunc) IRoutes {
	if matched := regEnLetter.MatchString(httpMethod); !matched {
		panic("http method " + httpMethod + " is not valid")
	}
	return group.handle(httpMethod, relativePath, handlers)
}

// GET is a shortcut for router.Handle("GET", path, handlers).
func (group *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle(http.MethodGet, relativePath, handlers)
}

func (group *RouterGroup) returnObj() IRoutes {
	return group
}
