package routes

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/middlewares"

	"github.com/gin-gonic/gin"
)

type CustomerRoute struct {
	handler *handler.CustomerHandler
	cfg     *config.Config
}

func NewCustomerRoute(h *handler.CustomerHandler, cfg *config.Config) *CustomerRoute {
	return &CustomerRoute{handler: h, cfg: cfg}
}

func (r *CustomerRoute) Register(v1 *gin.RouterGroup, admin *gin.RouterGroup) {
	// Public Membership lookup
	v1.GET("/customers/lookup", r.handler.LookupByPhone)

	// Admin customer list & manual adjustment (Owner)
	adminCust := admin.Group("/customers")
	adminCust.Use(middlewares.AuthMiddleware(r.cfg), middlewares.RequireOwner())
	{
		adminCust.GET("", r.handler.GetCustomers)
		adminCust.POST("/:id/adjust-points", r.handler.AdjustPoints)
	}
}
