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

	"go.mongodb.org/mongo-driver/v2/bson"
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
		_ = newAdmin.SetPIN("123456")
		_ = staffRepo.Create(ctx, newAdmin)
	} else if !admin.CheckPIN("123456") {
		// Đảm bảo mật khẩu mặc định được hash đúng bcrypt nếu trước đó chưa có hoặc hash lỗi
		_ = admin.SetPIN("123456")
		_, _ = staffRepo.Update(ctx, admin.ID, bson.M{"pinHash": admin.PINHash})
	}

	thungan, _ := staffRepo.FindByUsername(ctx, "thungan")
	if thungan == nil {
		newCashier := &model.StaffAccount{
			Username:   "thungan",
			Name:       "Trần Thu Ngân",
			Role:       model.RoleCashier,
			Shift:      "Ca Sáng (06:30 - 14:30)",
			HourlyRate: 28000,
			IsActive:   true,
		}
		_ = newCashier.SetPIN("123456")
		_ = staffRepo.Create(ctx, newCashier)
	} else if !thungan.CheckPIN("123456") {
		_ = thungan.SetPIN("123456")
		_, _ = staffRepo.Update(ctx, thungan.ID, bson.M{"pinHash": thungan.PINHash})
	}
}
