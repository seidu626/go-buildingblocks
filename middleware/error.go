package middleware

import (
	"github.com/valyala/fasthttp"
	"go.uber.org/zap"
)

func ApplyErrorMiddleware(next fasthttp.RequestHandler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Error handling middleware
		defer func() {
			if err := recover(); err != nil {
				logger := zap.L()
				logger.Error("Unhandled error", zap.Any("error", err))
				ctx.SetStatusCode(fasthttp.StatusInternalServerError)
				ctx.SetBody([]byte("Internal Server Error"))
			}
		}()

		next(ctx)
	}
}
