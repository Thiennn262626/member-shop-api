package routes

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/middlewares"

	"github.com/gin-gonic/gin"
)

type ProductRoute struct {
	handler *handler.ProductHandler
	cfg     *config.Config
}

func NewProductRoute(h *handler.ProductHandler, cfg *config.Config) *ProductRoute {
	return &ProductRoute{handler: h, cfg: cfg}
}

func (r *ProductRoute) Register(v1 *gin.RouterGroup, admin *gin.RouterGroup) {
	// 1. Public Menu
	v1.GET("/products", r.handler.GetProducts)
	v1.GET("/products/:id", r.handler.GetProductByID)

	// 2. Admin Stock Toggles
	stock := admin.Group("")
	stock.Use(middlewares.AuthMiddleware(r.cfg))
	{
		stock.PATCH("/products/:id/toggle", r.handler.ToggleProductStatus)
		stock.PATCH("/stock/options/:optionId", r.handler.ToggleOptionAvailability)
	}

	// 3. Owner CRUD
	adminOwner := admin.Group("")
	adminOwner.Use(middlewares.AuthMiddleware(r.cfg), middlewares.RequireOwner())
	{
		adminOwner.POST("/products", r.handler.CreateProduct)
		adminOwner.PUT("/products/:id", r.handler.UpdateProduct)
		adminOwner.DELETE("/products/:id", r.handler.DeleteProduct)
		adminOwner.PATCH("/stock/toggle-all", r.handler.ToggleAllProductsStatus)
	}
}
