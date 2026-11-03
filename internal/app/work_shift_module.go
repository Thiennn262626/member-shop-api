package app

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/repository"
	"member-shop-api/internal/routes"
)

type WorkShiftModule struct {
	handler *handler.WorkShiftHandler
	repo    repository.WorkShiftRepository
	cfg     *config.Config
}

func NewWorkShiftModule(mc *ModuleContext) Module {
	repo := repository.NewWorkShiftRepository(mc.DB)
	h := handler.NewWorkShiftHandler(repo, mc.StaffRepo)
	return &WorkShiftModule{
		handler: h,
		repo:    repo,
		cfg:     mc.Config,
	}
}

func (m *WorkShiftModule) Routes() routes.Route {
	return routes.NewWorkShiftRoute(m.handler, m.cfg)
}
