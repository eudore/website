package context

import (
	"github.com/eudore/eudore"
)

type GetContext = eudore.Context

type Context struct {
	GetContext
}

func init() {
	eudore.RegisterHandlerFunc(func(fn func(Context)) eudore.HandlerFunc {
		return func(ctx eudore.Context) {
			fn(Context{ctx})
		}
	})
}

// func (ctx *Context) GetContext() eudore.Context {
// 	return ctx.Context
// }
