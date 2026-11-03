package routes

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/middlewares"

	"github.com/gin-gonic/gin"
)

type AuthRoute struct {
	handler *handler.AuthHandler
	cfg     *config.Config
}

func NewAuthRoute(h *handler.AuthHandler, cfg *config.Config) *AuthRoute {
	return &AuthRoute{handler: h, cfg: cfg}
}

func (r *AuthRoute) Register(v1 *gin.RouterGroup, admin *gin.RouterGroup) {
	// Public
	v1.POST("/auth/login", r.handler.Login)

	// Protected Auth
	v1.GET("/auth/me", middlewares.AuthMiddleware(r.cfg), r.handler.GetMe)

	// Admin staff management (Owner only)
	staff := admin.Group("/staff")
	staff.Use(middlewares.AuthMiddleware(r.cfg), middlewares.RequireOwner())
	{
		staff.GET("", r.handler.GetAllStaff)
		staff.POST("", r.handler.CreateStaff)
		staff.PUT("/:id", r.handler.UpdateStaff)
		staff.PATCH("/:id/toggle", r.handler.ToggleStaffStatus)
	}
}
