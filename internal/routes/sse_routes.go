package routes

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/middlewares"

	"github.com/gin-gonic/gin"
)

type SSERoute struct {
	handler *handler.SSEHandler
	cfg     *config.Config
}

func NewSSERoute(h *handler.SSEHandler, cfg *config.Config) *SSERoute {
	return &SSERoute{handler: h, cfg: cfg}
}

func (r *SSERoute) Register(v1 *gin.RouterGroup, admin *gin.RouterGroup) {
	// SSE Realtime bảo vệ chỉ dành cho Admin / Thu ngân POS có Token xác thực
	admin.GET("/sse/events", middlewares.AuthMiddleware(r.cfg), r.handler.StreamEvents)
}
