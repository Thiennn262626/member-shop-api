package routes

import (
	"member-shop-api/internal/handler"

	"github.com/gin-gonic/gin"
)

type SSERoute struct {
	handler *handler.SSEHandler
}

func NewSSERoute(h *handler.SSEHandler) *SSERoute {
	return &SSERoute{handler: h}
}

func (r *SSERoute) Register(v1 *gin.RouterGroup, admin *gin.RouterGroup) {
	v1.GET("/sse/events", r.handler.StreamEvents)
}
