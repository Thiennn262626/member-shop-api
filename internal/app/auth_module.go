package app

import (
	"context"
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/model"
	"member-shop-api/internal/repository"
	"member-shop-api/internal/routes"
	"member-shop-api/internal/services"
	"time"
)

type AuthModule struct {
	handler *handler.AuthHandler
	service services.AuthService
	cfg     *config.Config
}

func NewAuthModule(mc *ModuleContext) Module {
	authSvc := mc.AuthSvc
	if authSvc == nil {
		authSvc = services.NewAuthService(mc.Config, mc.StaffRepo, mc.Redis)
		mc.AuthSvc = authSvc
	}

	h := handler.NewAuthHandler(authSvc, mc.StaffRepo, mc.Config)
	seedDefaultAdmin(mc.StaffRepo)

	return &AuthModule{
		handler: h,
		service: authSvc,
		cfg:     mc.Config,
	}
}

func (m *AuthModule) Routes() routes.Route {
	return routes.NewAuthRoute(m.handler, m.cfg)
}

func seedDefaultAdmin(staffRepo repository.StaffRepository) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	admin, _ := staffRepo.FindByUsername(ctx, "admin")
	if admin == nil {
		newAdmin := &model.StaffAccount{
			Username:   "admin",
			Name:       "Nguyễn Quản Lý",
			Role:       model.RoleOwner,
			HourlyRate: 50000,
			IsActive:   true,
		}
		_ = newAdmin.SetPIN("1234")
		_ = staffRepo.Create(ctx, newAdmin)
	}
}
