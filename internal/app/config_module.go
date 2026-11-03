package app

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/routes"
)

type ConfigModule struct {
	handler *handler.ConfigHandler
	cfg     *config.Config
}

func NewConfigModule(mc *ModuleContext) Module {
	h := handler.NewConfigHandler(mc.ConfigRepo)
	return &ConfigModule{
		handler: h,
		cfg:     mc.Config,
	}
}

func (m *ConfigModule) Routes() routes.Route {
	return routes.NewConfigRoute(m.handler, m.cfg)
}
