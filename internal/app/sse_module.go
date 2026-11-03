package app

import (
	"member-shop-api/internal/handler"
	"member-shop-api/internal/routes"
)

type SSEModule struct {
	handler *handler.SSEHandler
}

func NewSSEModule(mc *ModuleContext) Module {
	h := handler.NewSSEHandler(mc.SSEHub)
	return &SSEModule{
		handler: h,
	}
}

func (m *SSEModule) Routes() routes.Route {
	return routes.NewSSERoute(m.handler)
}
