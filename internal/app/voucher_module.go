package app

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/routes"
)

type VoucherModule struct {
	handler *handler.VoucherHandler
	cfg     *config.Config
}

func NewVoucherModule(mc *ModuleContext) Module {
	h := handler.NewVoucherHandler(mc.VoucherRepo, mc.Redis)
	return &VoucherModule{
		handler: h,
		cfg:     mc.Config,
	}
}

func (m *VoucherModule) Routes() routes.Route {
	return routes.NewVoucherRoute(m.handler, m.cfg)
}
