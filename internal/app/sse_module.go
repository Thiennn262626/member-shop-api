package app

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/routes"
)

type SSEModule struct {
	handler *handler.SSEHandler
	cfg     *config.Config
}

func NewSSEModule(mc *ModuleContext) Module {
	h := handler.NewSSEHandler(mc.SSEHub)
	return &SSEModule{
		handler: h,
		cfg:     mc.Config,
	}
}

func (m *SSEModule) Routes() routes.Route {
	return routes.NewSSERoute(m.handler, m.cfg)
}
