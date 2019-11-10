package middleware

import (
	"github.com/eudore/eudore"
	"time"
)

// NewLoggerFunc 函数创建一个请求日志记录中间件,不使用WithFields，在LoggerStd实现中会强制指定Fields顺序。
func NewLoggerFunc(app *eudore.App, params ...string) eudore.HandlerFunc {
	return func(ctx eudore.Context) {
		now := time.Now()
		ctx.Next()
		status := ctx.Response().Status()
		out := app.WithField("method", ctx.Method()).WithField("path", ctx.Path()).WithField("remote", ctx.RealIP()).WithField("proto", ctx.Request().Proto).WithField("host", ctx.Host()).WithField("status", status).WithField("time", time.Now().Sub(now).String()).WithField("size", ctx.Response().Size())

		for _, param := range params {
			val := ctx.GetParam(param)
			if val != "" {
				out = out.WithField(param, val)
			}
		}

		if requestID := ctx.GetHeader(eudore.HeaderXRequestID); len(requestID) > 0 {
			out = out.WithField("x-request-id", requestID)
		}
		if parentID := ctx.GetHeader(eudore.HeaderXParentID); len(parentID) > 0 {
			out = out.WithField("x-parent-id", parentID)
		}

		if 300 < status && status < 400 && status != 304 {
			out = out.WithField("location", ctx.Response().Header().Get(eudore.HeaderLocation))
		}
		if status < 400 {
			out.Info()
		} else {
			if err := ctx.Err(); err != nil {
				out = out.WithField("error", err.Error())
			}
			out.Error()
		}
	}
}
