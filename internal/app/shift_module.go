package app

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/routes"
)

type ShiftModule struct {
	handler *handler.ShiftHandler
	cfg     *config.Config
}

func NewShiftModule(mc *ModuleContext) Module {
	h := handler.NewShiftHandler(mc.ShiftRepo)
	return &ShiftModule{
		handler: h,
		cfg:     mc.Config,
	}
}

func (m *ShiftModule) Routes() routes.Route {
	return routes.NewShiftRoute(m.handler, m.cfg)
}
