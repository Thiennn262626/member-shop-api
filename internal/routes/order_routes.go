package routes

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/middlewares"

	"github.com/gin-gonic/gin"
)

type OrderRoute struct {
	handler *handler.OrderHandler
	cfg     *config.Config
}

func NewOrderRoute(h *handler.OrderHandler, cfg *config.Config) *OrderRoute {
	return &OrderRoute{handler: h, cfg: cfg}
}

func (r *OrderRoute) Register(v1 *gin.RouterGroup, admin *gin.RouterGroup) {
	// Public
	v1.POST("/orders", r.handler.CreateWebOrder)
	v1.GET("/orders/track", r.handler.TrackOrders)

	// Protected POS order creation
	v1.POST("/pos/orders", middlewares.AuthMiddleware(r.cfg), r.handler.CreatePosOrder)

	// Admin Orders
	orders := admin.Group("/orders")
	orders.Use(middlewares.AuthMiddleware(r.cfg))
	{
		orders.GET("", r.handler.GetOrders)
		orders.GET("/:id", r.handler.GetOrderByID)
		orders.PATCH("/:id/status", r.handler.UpdateOrderStatus)
	}
}
