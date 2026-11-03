package app

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/middlewares"
	"member-shop-api/internal/repository"
	"member-shop-api/internal/services"
	"member-shop-api/pkg/cache"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// ModuleContext encapsulates shared dependencies and context passed down to all feature modules.
type ModuleContext struct {
	Config *config.Config
	DB     *mongo.Database
	Redis  cache.RedisCacheService
	Engine *gin.Engine
	SSEHub *services.SSEHub

	// Shared cross-module repositories
	StaffRepo   repository.StaffRepository
	ConfigRepo  repository.ConfigRepository
	CustRepo    repository.CustomerRepository
	PointTxRepo repository.PointTransactionRepository
	ProductRepo repository.ProductRepository
	VoucherRepo repository.VoucherRepository
	ShiftRepo   repository.ShiftRepository
	OrderRepo   repository.OrderRepository

	// Shared cross-module services
	OrderSvc services.OrderService
	AuthSvc  services.AuthService
}

func NewModuleContext(
	cfg *config.Config,
	db *mongo.Database,
	rdb cache.RedisCacheService,
	engine *gin.Engine,
	sseHub *services.SSEHub,
) *ModuleContext {
	staffRepo := repository.NewStaffRepository(db)
	configRepo := repository.NewConfigRepository(db)
	custRepo := repository.NewCustomerRepository(db)
	pointTxRepo := repository.NewPointTransactionRepository(db)
	productRepo := repository.NewProductRepository(db)
	voucherRepo := repository.NewVoucherRepository(db)
	shiftRepo := repository.NewShiftRepository(db)
	orderRepo := repository.NewOrderRepository(db)

	authSvc := services.NewAuthService(cfg, staffRepo, rdb)
	orderSvc := services.NewOrderService(
		orderRepo, custRepo, productRepo, voucherRepo, shiftRepo, configRepo, pointTxRepo, sseHub, rdb,
	)

	// Khởi tạo Middleware Auth với ECDSA Key Manager và Cache Blacklist
	middlewares.InitAuthMiddleware(authSvc.GetKeyManager(), rdb)

	return &ModuleContext{
		Config:      cfg,
		DB:          db,
		Redis:       rdb,
		Engine:      engine,
		SSEHub:      sseHub,
		StaffRepo:   staffRepo,
		ConfigRepo:  configRepo,
		CustRepo:    custRepo,
		PointTxRepo: pointTxRepo,
		ProductRepo: productRepo,
		VoucherRepo: voucherRepo,
		ShiftRepo:   shiftRepo,
		OrderRepo:   orderRepo,
		OrderSvc:    orderSvc,
		AuthSvc:     authSvc,
	}
}
