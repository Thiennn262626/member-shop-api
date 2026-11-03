package app

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/routes"
)

type CustomerModule struct {
	handler *handler.CustomerHandler
	cfg     *config.Config
}

func NewCustomerModule(mc *ModuleContext) Module {
	h := handler.NewCustomerHandler(mc.CustRepo, mc.PointTxRepo)
	return &CustomerModule{
		handler: h,
		cfg:     mc.Config,
	}
}

func (m *CustomerModule) Routes() routes.Route {
	return routes.NewCustomerRoute(m.handler, m.cfg)
}
