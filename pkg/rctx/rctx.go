package rctx

import (
	"context"
	"errors"
	"net/http"
	"time"
)

type Context struct {
	parent context.Context
	params *routeParams
	err    error
}

// PrepareRequestContext prepares the context of a request for matching.
func PrepareRequestContext(req *http.Request, maxParams int) *http.Request {
	rctx := &Context{
		parent: req.Context(),
		params: newParams(maxParams),
		err:    nil,
	}
	return req.WithContext(rctx)
}

// ResetRequestContext resets any values in the context that shouldn't be maintained between attempts to match routes.
// This assumes that the request has a rctx.Context, and returns an error if it does not.
func ResetRequestContext(req *http.Request) error {
	ctx := req.Context()
	rctx, correctType := ctx.(*Context)
	if !correctType {
		return errors.New("placeholder error; request must have *rctx.Context when resetting")
	}
	rctx.params.head = 0
	return nil
}

// PARAMETER IMPLEMENTATION

// GetParam gets a parameter by its key string.
// This automatically converts key to its underlying context key type.
func GetParam(key string, ctx *Context) string {
	return ctx.params.get(paramKey(key))
}

// SetParam sets a parameter with a key string.
// This automatically converts key to its underlying context key type.
// Context params have a max value determined at creation, and this returns an error if the user attempts to exceed
// the maximum number of params.
func SetParam(key, value string, ctx *Context) error {
	return ctx.params.set(paramKey(key), value)
}

// CONTEXT IMPLEMENTATION

// rctx.Context does not natively support deadlines.
// If the parent context has a deadline, that will be returned.
//
// See interface context.Context.
func (ctx *Context) Deadline() (time.Time, bool) {
	if ctx.parent != nil {
		return ctx.parent.Deadline()
	}
	return time.Time{}, false
}

// rctx.Context does not natively support doneness signals.
// If the parent context has a doneness signal, that will be returned.
//
// See interface context.Context.
func (ctx *Context) Done() <-chan struct{} {
	if ctx.parent != nil {
		return ctx.parent.Done()
	}
	return nil
}

// Return the current error on the context, or the error on the parent if applicable.
//
// See interface context.Context.
func (ctx *Context) Err() error {
	if ctx.err != nil {
		return ctx.err
	} else if ctx.parent != nil {
		return ctx.parent.Err()
	}
	return nil
}

// Get a value.
//
// See interface context.Context.
func (ctx *Context) Value(key any) any {
	if pkey, ok := key.(paramKey); ok {
		return ctx.params.get(pkey)
	} else {
		return ctx.parent.Value(key)
	}
}
