package app

import (
	"context"
	"fmt"
	"member-shop-api/internal/config"
	"member-shop-api/internal/db"
	"member-shop-api/internal/routes"
	"member-shop-api/internal/services"
	"member-shop-api/internal/utils"
	"member-shop-api/pkg/cache"
	"member-shop-api/pkg/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type Module interface {
	Routes() routes.Route
}

type Application struct {
	cfg    *config.Config
	mongo  *db.MongoDB
	server *http.Server
}

func getModuleRoutes(modules []Module) []routes.Route {
	var r []routes.Route
	for _, m := range modules {
		r = append(r, m.Routes())
	}
	return r
}

func NewApplication(cfg *config.Config) (*Application, error) {
	// 1. Kết nối MongoDB
	mongoDB, err := db.ConnectMongoDB(cfg)
	if err != nil {
		return nil, fmt.Errorf("connect mongodb failed: %w", err)
	}

	// 2. Cấu hình chế độ Gin
	if cfg.AppEnv == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	engine := gin.New()

	// Vô hiệu hóa việc tin tưởng các proxy không xác định để chống IP Spoofing qua header X-Forwarded-For (SEC-05)
	_ = engine.SetTrustedProxies(nil)

	// 3. Khởi tạo Cache Service (Ưu tiên Redis, tự động fallback sang High-performance In-Memory Cache)
	var cacheRedisService cache.RedisCacheService
	redisClient := config.NewRedisClient()
	if redisClient != nil {
		redisLogger := utils.NewLoggerWithPath("redis.log", zerolog.InfoLevel)
		cacheRedisService = cache.NewRedisCacheService(redisClient, redisLogger)
		logger.Log.Info().Msg("🚀 Redis Cache Service activated")
	} else {
		cacheRedisService = cache.NewMemoryCacheService()
		logger.Log.Info().Msg("⚡ In-Memory Cache Service activated (Cache-Aside ready)")
	}

	// 4. Khởi tạo ModuleContext chứa các thành phần dùng chung truyền xuống các module
	mc := NewModuleContext(
		cfg,
		mongoDB.Database,
		cacheRedisService,
		engine,
		services.GlobalSSEHub,
	)

	// 5. Khởi tạo các Modules chuẩn theo setup của dự án
	modules := []Module{
		NewAuthModule(mc),
		NewConfigModule(mc),
		NewProductModule(mc),
		NewVoucherModule(mc),
		NewShiftModule(mc),
		NewCustomerModule(mc),
		NewOrderModule(mc),
		NewWorkShiftModule(mc),
		NewSSEModule(mc),
	}

	// 6. Đăng ký toàn bộ Middlewares & Routes qua hàm routes.RegisterRoutes
	routes.RegisterRoutes(engine, getModuleRoutes(modules)...)

	// 7. Seed dữ liệu khởi tạo nếu database trống (Sản phẩm, Voucher, Hội viên mẫu)
	SeedInitialData(mc)

	server := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: engine,
	}

	return &Application{
		cfg:    cfg,
		mongo:  mongoDB,
		server: server,
	}, nil
}

func (a *Application) Run(ctx context.Context) error {
	logger.Log.Info().Str("port", a.cfg.Port).Msg("🚀 Member Shop API server is starting...")

	errChan := make(chan error, 1)
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- err
		}
	}()

	select {
	case err := <-errChan:
		return err
	case <-ctx.Done():
		logger.Log.Info().Msg("Shutting down API server gracefully...")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := a.server.Shutdown(shutdownCtx); err != nil {
			logger.Log.Error().Err(err).Msg("Server forced shutdown error")
		}

		if err := a.mongo.Close(shutdownCtx); err != nil {
			logger.Log.Error().Err(err).Msg("MongoDB disconnect error")
		}

		logger.Log.Info().Msg("Server stopped safely")
		return nil
	}
}
