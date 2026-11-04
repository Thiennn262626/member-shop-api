package handler

import (
	"math"
	"member-shop-api/internal/model"
	"member-shop-api/internal/repository"
	"member-shop-api/internal/utils"
	"member-shop-api/pkg/apperror"
	"member-shop-api/pkg/cache"
	"member-shop-api/pkg/response"
	"net/http"
	"strconv"
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
	hasPagination := c.Query("page") != "" || c.Query("pageSize") != "" || c.Query("limit") != "" || c.Query("search") != "" || c.Query("category") != ""

	// Nếu yêu cầu phân trang hoặc lọc
	if hasPagination {
		page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 64)
		if page < 1 {
			page = 1
		}

		limit, _ := strconv.ParseInt(c.DefaultQuery("pageSize", c.DefaultQuery("limit", "20")), 10, 64)
		if limit <= 0 {
			limit = 20
		} else if limit > 200 {
			limit = 200
		}

		var offset int64
		if offsetStr := c.Query("offset"); offsetStr != "" {
			offset, _ = strconv.ParseInt(offsetStr, 10, 64)
			if offset < 0 {
				offset = 0
			}
			page = (offset / limit) + 1
		} else {
			offset = (page - 1) * limit
		}

		filter := repository.ProductFilter{
			Category:   c.Query("category"),
			Search:     c.Query("search"),
			OnlyActive: onlyActive,
			Limit:      limit,
			Offset:     offset,
		}

		products, total, err := h.repo.GetList(c.Request.Context(), filter)
		if err != nil {
			response.HandleError(c, err)
			return
		}
		if products == nil {
			products = []model.Product{}
		}

		totalPages := int(math.Ceil(float64(total) / float64(limit)))
		if totalPages == 0 && total == 0 {
			totalPages = 1
		}

		response.SuccessWithMeta(c, products, &response.Meta{
			Page:       int(page),
			PageSize:   int(limit),
			Total:      total,
			TotalPages: totalPages,
		})
		return
	}

	// Trường hợp lấy toàn bộ menu (cho Landing Page & POS) kèm Cache-Aside
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
	if products == nil {
		products = []model.Product{}
	}

	if onlyActive {
		c.Header("Cache-Control", "public, max-age=60, stale-while-revalidate=120")
	}

	total := int64(len(products))
	response.SuccessWithMeta(c, products, &response.Meta{
		Page:       1,
		PageSize:   int(total),
		Total:      total,
		TotalPages: 1,
	})
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

	expectedVersion := utils.ParseExpectedVersion(updates)
	delete(updates, "_id")
	delete(updates, "id")

	var updated *model.Product
	if expectedVersion != nil {
		updated, err = h.repo.Update(c.Request.Context(), id, updates, *expectedVersion)
	} else {
		updated, err = h.repo.Update(c.Request.Context(), id, updates)
	}
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
