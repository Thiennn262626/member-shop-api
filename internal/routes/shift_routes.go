package routes

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/middlewares"

	"github.com/gin-gonic/gin"
)

type ShiftRoute struct {
	handler *handler.ShiftHandler
	cfg     *config.Config
}

func NewShiftRoute(h *handler.ShiftHandler, cfg *config.Config) *ShiftRoute {
	return &ShiftRoute{handler: h, cfg: cfg}
}

func (r *ShiftRoute) Register(v1 *gin.RouterGroup, admin *gin.RouterGroup) {
	shifts := admin.Group("/shifts")
	shifts.Use(middlewares.AuthMiddleware(r.cfg))
	{
		shifts.GET("/current", r.handler.GetCurrentShift)
		shifts.POST("/open", r.handler.OpenShift)
		shifts.POST("/:id/close", r.handler.CloseShift)
		shifts.GET("/history", middlewares.RequireOwner(), r.handler.GetHistory)
	}
}
