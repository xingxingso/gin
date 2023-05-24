// Copyright 2014 Manu Martinez-Almeida. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package gin

import "net/http"

const (
	noWritten     = -1
	defaultStatus = http.StatusOK
)

// ResponseWriter ...
type ResponseWriter interface {
	http.ResponseWriter

	// Status returns the HTTP response status code of the current request.
	Status() int
}

type responseWriter struct {
	http.ResponseWriter
	size   int
	status int
}

var _ ResponseWriter = (*responseWriter)(nil)

func (w *responseWriter) reset(writer http.ResponseWriter) {
	w.ResponseWriter = writer
	w.size = noWritten
	w.status = defaultStatus
}

func (w *responseWriter) WriteHeader(code int) {
	if code > 0 && w.status != code {
		if w.Written() {
			debugPrint("[WARNING] Headers were already written. Wanted to override status code %d with %d", w.status, code)
			return
		}
		w.status = code
	}
}

func (w *responseWriter) WriteHeaderNow() {
	if !w.Written() {
		w.size = 0
		w.ResponseWriter.WriteHeader(w.status)
	}
}

func (w *responseWriter) Write(data []byte) (n int, err error) {
	w.WriteHeaderNow()
	n, err = w.ResponseWriter.Write(data)
	w.size += n
	return
}

func (w *responseWriter) Status() int {
	return w.status
}

func (w *responseWriter) Written() bool {
	return w.size != noWritten
}
