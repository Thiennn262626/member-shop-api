package app

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/repository"
	"member-shop-api/internal/routes"
)

type ProductModule struct {
	handler *handler.ProductHandler
	repo    repository.ProductRepository
	cfg     *config.Config
}

func NewProductModule(mc *ModuleContext) Module {
	h := handler.NewProductHandler(mc.ProductRepo)
	return &ProductModule{
		handler: h,
		repo:    mc.ProductRepo,
		cfg:     mc.Config,
	}
}

func (m *ProductModule) Routes() routes.Route {
	return routes.NewProductRoute(m.handler, m.cfg)
}
