package handler

import (
	"member-shop-api/internal/model"
	"member-shop-api/internal/repository"
	"member-shop-api/pkg/apperror"
	"member-shop-api/pkg/cache"
	"member-shop-api/pkg/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ProductHandler struct {
	repo  repository.ProductRepository
	cache cache.RedisCacheService
}

func NewProductHandler(repo repository.ProductRepository, cacheSvc cache.RedisCacheService) *ProductHandler {
	return &ProductHandler{repo: repo, cache: cacheSvc}
}

func (h *ProductHandler) GetProducts(c *gin.Context) {
	onlyActive := c.Query("all") != "true"
	cacheKey := cache.KeyProductsActive
	ttl := 15 * time.Minute
	if !onlyActive {
		cacheKey = cache.KeyProductsAll
		ttl = 5 * time.Minute
	}

	products, err := cache.CacheAside(c.Request.Context(), h.cache, cacheKey, ttl, func() ([]model.Product, error) {
		return h.repo.GetAll(c.Request.Context(), onlyActive)
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, products)
}

func (h *ProductHandler) GetProductByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, apperror.CodeBadRequest, "Mã sản phẩm không đúng định dạng")
		return
	}

	cacheKey := "shop:product:" + idStr
	prod, err := cache.CacheAside(c.Request.Context(), h.cache, cacheKey, 30*time.Minute, func() (*model.Product, error) {
		return h.repo.GetByID(c.Request.Context(), id)
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}
	if prod == nil {
		response.FailWithStatus(c, http.StatusNotFound, apperror.CodeProductNotFound, "Không tìm thấy thông tin sản phẩm")
		return
	}
	response.Success(c, prod)
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	var p model.Product
	if err := c.ShouldBindJSON(&p); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	if err := h.repo.Create(c.Request.Context(), &p); err != nil {
		response.HandleError(c, err)
		return
	}

	// Invalidate Cache liên quan đến sản phẩm
	cache.InvalidatePattern(c.Request.Context(), h.cache, cache.PatternProducts)

	response.CreatedWithMsg(c, "Tạo sản phẩm thành công", p)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, apperror.CodeBadRequest, "Mã sản phẩm không đúng định dạng")
		return
	}

	var updates bson.M
	if err := c.ShouldBindJSON(&updates); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	delete(updates, "_id")
	delete(updates, "id")

	updated, err := h.repo.Update(c.Request.Context(), id, updates)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	// Invalidate Cache liên quan đến sản phẩm
	cache.InvalidatePattern(c.Request.Context(), h.cache, cache.PatternProducts)

	response.SuccessWithMsg(c, "Cập nhật sản phẩm thành công", updated)
}

func (h *ProductHandler) ToggleProductStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, apperror.CodeBadRequest, "Mã sản phẩm không đúng định dạng")
		return
	}

	updated, err := h.repo.ToggleActive(c.Request.Context(), id)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	// Invalidate Cache liên quan đến sản phẩm
	cache.InvalidatePattern(c.Request.Context(), h.cache, cache.PatternProducts)

	response.SuccessWithMsg(c, "Cập nhật trạng thái sản phẩm thành công", updated)
}

func (h *ProductHandler) ToggleAllProductsStatus(c *gin.Context) {
	var req struct {
		IsActive bool   `json:"isActive"`
		Category string `json:"category"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	if err := h.repo.ToggleAllActive(c.Request.Context(), req.IsActive, req.Category); err != nil {
		response.HandleError(c, err)
		return
	}

	// Invalidate Cache liên quan đến sản phẩm
	cache.InvalidatePattern(c.Request.Context(), h.cache, cache.PatternProducts)

	response.SuccessWithMsg(c, "Thay đổi trạng thái danh mục sản phẩm thành công", nil)
}

func (h *ProductHandler) ToggleOptionAvailability(c *gin.Context) {
	optionID := c.Param("optionId")
	var req struct {
		IsAvailable bool `json:"isAvailable"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	if err := h.repo.ToggleOptionAvailability(c.Request.Context(), optionID, req.IsAvailable); err != nil {
		response.HandleError(c, err)
		return
	}

	// Invalidate Cache liên quan đến sản phẩm
	cache.InvalidatePattern(c.Request.Context(), h.cache, cache.PatternProducts)

	response.SuccessWithMsg(c, "Thay đổi trạng thái tùy chọn thành công", nil)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, apperror.CodeBadRequest, "Mã sản phẩm không đúng định dạng")
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		response.HandleError(c, err)
		return
	}

	// Invalidate Cache liên quan đến sản phẩm
	cache.InvalidatePattern(c.Request.Context(), h.cache, cache.PatternProducts)

	response.SuccessWithMsg(c, "Xóa sản phẩm thành công", nil)
}
