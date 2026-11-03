package app

import (
	"member-shop-api/internal/config"
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
	return &ModuleContext{
		Config:      cfg,
		DB:          db,
		Redis:       rdb,
		Engine:      engine,
		SSEHub:      sseHub,
		StaffRepo:   repository.NewStaffRepository(db),
		ConfigRepo:  repository.NewConfigRepository(db),
		CustRepo:    repository.NewCustomerRepository(db),
		PointTxRepo: repository.NewPointTransactionRepository(db),
		ProductRepo: repository.NewProductRepository(db),
		VoucherRepo: repository.NewVoucherRepository(db),
		ShiftRepo:   repository.NewShiftRepository(db),
		OrderRepo:   repository.NewOrderRepository(db),
	}
}
