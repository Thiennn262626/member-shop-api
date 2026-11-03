package middlewares

import (
	"context"
	"member-shop-api/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TraceMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		traceID := ctx.GetHeader("X-Trace-Id")
		if traceID == "" {
			traceID = uuid.New().String()
		}

		contextValue := context.WithValue(ctx.Request.Context(), logger.TraceIdKey, traceID)
		ctx.Request = ctx.Request.WithContext(contextValue)

		ctx.Writer.Header().Set("X-Trace-Id", traceID)

		ctx.Set(string(logger.TraceIdKey), traceID)

		ctx.Next()
	}
}
