package middlewares

import (
	"bytes"
	"fmt"
	"member-shop-api/pkg/apperror"
	"member-shop-api/pkg/logger"
	"member-shop-api/pkg/response"
	"net/http"
	"regexp"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func RecoveryMiddleware(recoveryLogger *zerolog.Logger) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				stack := debug.Stack()

				stack_at := ExtractFirstAppStackLine(stack)

				recoveryLogger.Error().
					Str("trace_id", logger.GetTraceID(ctx.Request.Context())).
					Str("path", ctx.Request.URL.Path).
					Str("method", ctx.Request.Method).
					Str("client_ip", ctx.ClientIP()).
					Str("panic", fmt.Sprintf("%v", err)).
					Str("statck_at", stack_at).
					Str("statck", string(stack)).
					Msg("panic occurred")

				response.FailWithStatus(ctx, http.StatusInternalServerError, apperror.CodeInternalServerError, "Hệ thống gặp sự cố ngoài dự kiến, vui lòng thử lại sau.")
				ctx.Abort()
			}
		}()

		ctx.Next()
	}
}

var stackLineRegex = regexp.MustCompile(`(.+\.go:\d+)`)

func ExtractFirstAppStackLine(stack []byte) string {
	lines := bytes.Split(stack, []byte("\n"))

	for _, line := range lines {
		if bytes.Contains(line, []byte(".go")) &&
			!bytes.Contains(line, []byte("/runtime/")) &&
			!bytes.Contains(line, []byte("/debug")) &&
			!bytes.Contains(line, []byte("recovery_middleware.go")) {
			cleanLine := strings.TrimSpace(string(line))
			math := stackLineRegex.FindStringSubmatch(cleanLine)

			if len(math) > 1 {
				return math[1]
			}
		}
	}
	return ""
}
