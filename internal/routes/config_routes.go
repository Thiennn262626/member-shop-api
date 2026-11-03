package routes

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/middlewares"

	"github.com/gin-gonic/gin"
)

type ConfigRoute struct {
	handler *handler.ConfigHandler
	cfg     *config.Config
}

func NewConfigRoute(h *handler.ConfigHandler, cfg *config.Config) *ConfigRoute {
	return &ConfigRoute{handler: h, cfg: cfg}
}

func (r *ConfigRoute) Register(v1 *gin.RouterGroup, admin *gin.RouterGroup) {
	// Public
	v1.GET("/config", r.handler.GetConfig)

	// Owner Admin
	admin.PUT("/config", middlewares.AuthMiddleware(r.cfg), middlewares.RequireOwner(), r.handler.UpdateConfig)
}
