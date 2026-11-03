package routes

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/middlewares"

	"github.com/gin-gonic/gin"
)

type VoucherRoute struct {
	handler *handler.VoucherHandler
	cfg     *config.Config
}

func NewVoucherRoute(h *handler.VoucherHandler, cfg *config.Config) *VoucherRoute {
	return &VoucherRoute{handler: h, cfg: cfg}
}

func (r *VoucherRoute) Register(v1 *gin.RouterGroup, admin *gin.RouterGroup) {
	// Public
	v1.POST("/vouchers/validate", r.handler.ValidateVoucher)

	// Owner Admin
	vouchers := admin.Group("/vouchers")
	vouchers.Use(middlewares.AuthMiddleware(r.cfg), middlewares.RequireOwner())
	{
		vouchers.GET("", r.handler.GetAll)
		vouchers.POST("", r.handler.Create)
		vouchers.PUT("/:id", r.handler.Update)
		vouchers.DELETE("/:id", r.handler.Delete)
	}
}
