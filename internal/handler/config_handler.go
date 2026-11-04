package handler

import (
	"member-shop-api/internal/model"
	"member-shop-api/internal/repository"
	"member-shop-api/internal/utils"
	"member-shop-api/pkg/cache"
	"member-shop-api/pkg/response"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ConfigHandler struct {
	repo  repository.ConfigRepository
	cache cache.RedisCacheService
}

func NewConfigHandler(repo repository.ConfigRepository, cacheSvc cache.RedisCacheService) *ConfigHandler {
	return &ConfigHandler{repo: repo, cache: cacheSvc}
}

func (h *ConfigHandler) GetConfig(c *gin.Context) {
	// Cache-Aside: Cấu hình cửa hàng đọc từ Cache (TTL: 1 giờ)
	cfg, err := cache.CacheAside(c.Request.Context(), h.cache, cache.KeyShopConfig, 1*time.Hour, func() (*model.ShopConfig, error) {
		return h.repo.GetConfig(c.Request.Context())
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}

	// Tối ưu hiệu năng HTTP Cache cho cấu hình cửa hàng (5 phút browser cache + 10 phút stale-while-revalidate)
	c.Header("Cache-Control", "public, max-age=300, stale-while-revalidate=600")
	response.Success(c, cfg)
}

func (h *ConfigHandler) UpdateConfig(c *gin.Context) {
	var updates bson.M
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	expectedVersion := utils.ParseExpectedVersion(updates)
	delete(updates, "_id")
	delete(updates, "id")

	var updated *model.ShopConfig
	var err error
	if expectedVersion != nil {
		updated, err = h.repo.UpdateConfig(c.Request.Context(), updates, *expectedVersion)
	} else {
		updated, err = h.repo.UpdateConfig(c.Request.Context(), updates)
	}
	if err != nil {
		response.HandleError(c, err)
		return
	}

	// Invalidate Cache ngay khi cấu hình thay đổi
	cache.Invalidate(c.Request.Context(), h.cache, cache.KeyShopConfig)

	response.SuccessWithMsg(c, "Cập nhật cấu hình cửa hàng thành công", updated)
}
