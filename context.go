package gin

import (
	"net/http"

	"github.com/xingxingso/gin/binding"
)

const (
	MIMEJSON              = binding.MIMEJSON
	MIMEHTML              = binding.MIMEHTML
	MIMEXML               = binding.MIMEXML
	MIMEXML2              = binding.MIMEXML2
	MIMEPlain             = binding.MIMEPlain
	MIMEPOSTForm          = binding.MIMEPOSTForm
	MIMEMultipartPOSTForm = binding.MIMEMultipartPOSTForm
	MIMEYAML              = binding.MIMEYAML
	MIMETOML              = binding.MIMETOML
)

// Context is the most important part of gin. It allows us to pass variables between middleware,
// manage the flow, validate the JSON of a request and render a JSON response for example.
type Context struct {
	writermem responseWriter
	Request   *http.Request
	Writer    ResponseWriter
}

/************************************/
/********** CONTEXT CREATION ********/
/************************************/

func (c *Context) reset() {
	c.Writer = &c.writermem
}

// JSON serializes the given struct as JSON into the response body.
// It also sets the Content-Type as "application/json".
func (c *Context) JSON(code int, obj any) {
	//c.Render(code, render.JSON{Data: obj})
}
