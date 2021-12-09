// Copyright 2021 The Gitea Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"code.gitea.io/gitea/modules/json"
	"code.gitea.io/gitea/modules/log"

	chi "github.com/go-chi/chi/v5"
)

// BaseContext represents a general context for some simple routes
type BaseContext struct {
	Resp ResponseWriter
	Req  *http.Request
	Data map[string]interface{}
}

// NewBaseContext creates a new base context
func NewBaseContext(resp http.ResponseWriter, req *http.Request, data map[string]interface{}) *BaseContext {
	return &BaseContext{
		Resp: NewResponse(resp),
		Req:  req,
		Data: data,
	}
}

// GetData returns the data
func (ctx *BaseContext) GetData() map[string]interface{} {
	return ctx.Data
}

// HasDataKey returns true if value of given name exists.
func (ctx *BaseContext) HasDataKey(name string) bool {
	_, ok := ctx.Data[name]
	return ok
}

// Header returns a header
func (ctx *BaseContext) Header() http.Header {
	return ctx.Resp.Header()
}

// RemoteAddr returns the client machie ip address
func (ctx *BaseContext) RemoteAddr() string {
	return ctx.Req.RemoteAddr
}

// Params returns the param on route
func (ctx *BaseContext) Params(p string) string {
	s, _ := url.PathUnescape(chi.URLParam(ctx.Req, strings.TrimPrefix(p, ":")))
	return s
}

// ParamsInt64 returns the param on route as int64
func (ctx *BaseContext) ParamsInt64(p string) int64 {
	v, _ := strconv.ParseInt(ctx.Params(p), 10, 64)
	return v
}

// SetParams set params into routes
func (ctx *BaseContext) SetParams(k, v string) {
	chiCtx := chi.RouteContext(ctx)
	chiCtx.URLParams.Add(strings.TrimPrefix(k, ":"), url.PathEscape(v))
}

// Write writes data to webbrowser
func (ctx *BaseContext) Write(bs []byte) (int, error) {
	return ctx.Resp.Write(bs)
}

// Written returns true if there are something sent to web browser
func (ctx *BaseContext) Written() bool {
	return ctx.Resp.Status() > 0
}

// Status writes status code
func (ctx *BaseContext) Status(status int) {
	ctx.Resp.WriteHeader(status)
}

// Deadline is part of the interface for context.Context and we pass this to the request context
func (ctx *BaseContext) Deadline() (deadline time.Time, ok bool) {
	return ctx.Req.Context().Deadline()
}

// Done is part of the interface for context.Context and we pass this to the request context
func (ctx *BaseContext) Done() <-chan struct{} {
	return ctx.Req.Context().Done()
}

// Err is part of the interface for context.Context and we pass this to the request context
func (ctx *BaseContext) Err() error {
	return ctx.Req.Context().Err()
}

// Value is part of the interface for context.Context and we pass this to the request context
func (ctx *BaseContext) Value(key interface{}) interface{} {
	return ctx.Req.Context().Value(key)
}

// Error returned an error to web browser
func (ctx *BaseContext) Error(status int, contents ...string) {
	var v = http.StatusText(status)
	if len(contents) > 0 {
		v = contents[0]
	}
	http.Error(ctx.Resp, v, status)
}

// Redirect redirect the request
func (ctx *BaseContext) Redirect(location string, status ...int) {
	code := http.StatusFound
	if len(status) == 1 {
		code = status[0]
	}

	http.Redirect(ctx.Resp, ctx.Req, location, code)
}

// JSON render content as JSON
func (ctx *BaseContext) JSON(status int, content interface{}) {
	ctx.Resp.Header().Set("Content-Type", "application/json;charset=utf-8")
	ctx.Resp.WriteHeader(status)
	if err := json.NewEncoder(ctx.Resp).Encode(content); err != nil {
		log.Error("Render JSON failed: %v", err)
		ctx.Status(http.StatusInternalServerError)
	}
}

// PlainText render content as plain text
func (ctx *BaseContext) PlainText(status int, bs []byte) {
	ctx.Resp.WriteHeader(status)
	ctx.Resp.Header().Set("Content-Type", "text/plain;charset=utf-8")
	if _, err := ctx.Resp.Write(bs); err != nil {
		log.Error("Render PlainText failed: %v", err)
		ctx.Status(http.StatusInternalServerError)
	}
}

// ServeFile serves given file to response.
func (ctx *BaseContext) ServeFile(file string, names ...string) {
	var name string
	if len(names) > 0 {
		name = names[0]
	} else {
		name = path.Base(file)
	}
	ctx.Resp.Header().Set("Content-Description", "File Transfer")
	ctx.Resp.Header().Set("Content-Type", "application/octet-stream")
	ctx.Resp.Header().Set("Content-Disposition", "attachment; filename="+name)
	ctx.Resp.Header().Set("Content-Transfer-Encoding", "binary")
	ctx.Resp.Header().Set("Expires", "0")
	ctx.Resp.Header().Set("Cache-Control", "must-revalidate")
	ctx.Resp.Header().Set("Pragma", "public")
	http.ServeFile(ctx.Resp, ctx.Req, file)
}
