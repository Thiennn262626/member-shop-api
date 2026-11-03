package app

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/repository"
	"member-shop-api/internal/routes"
	"member-shop-api/internal/services"
)

type OrderModule struct {
	handler *handler.OrderHandler
	service services.OrderService
	repo    repository.OrderRepository
	cfg     *config.Config
}

func NewOrderModule(mc *ModuleContext) Module {
	orderSvc := services.NewOrderService(
		mc.OrderRepo,
		mc.CustRepo,
		mc.ProductRepo,
		mc.VoucherRepo,
		mc.ShiftRepo,
		mc.ConfigRepo,
		mc.PointTxRepo,
		mc.SSEHub,
		mc.Redis,
	)
	mc.OrderSvc = orderSvc

	h := handler.NewOrderHandler(orderSvc, mc.OrderRepo, mc.Redis)

	return &OrderModule{
		handler: h,
		service: orderSvc,
		repo:    mc.OrderRepo,
		cfg:     mc.Config,
	}
}

func (m *OrderModule) Routes() routes.Route {
	return routes.NewOrderRoute(m.handler, m.cfg)
}
