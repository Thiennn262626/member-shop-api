package handler

import (
	"math"
	"member-shop-api/internal/dto"
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

type VoucherHandler struct {
	repo  repository.VoucherRepository
	cache cache.RedisCacheService
}

func NewVoucherHandler(repo repository.VoucherRepository, cacheSvc cache.RedisCacheService) *VoucherHandler {
	return &VoucherHandler{repo: repo, cache: cacheSvc}
}

func (h *VoucherHandler) ValidateVoucher(c *gin.Context) {
	var req dto.ValidateVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	// Cache-Aside: Đọc voucher từ Cache (TTL: 10 phút)
	vKey := cache.KeyVoucherPrefix + req.Code
	v, err := cache.CacheAside(c.Request.Context(), h.cache, vKey, 10*time.Minute, func() (*model.Voucher, error) {
		return h.repo.GetByCode(c.Request.Context(), req.Code)
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}
	if v == nil {
		response.Success(c, model.VoucherValidationResult{
			IsValid: false,
			Message: "Mã giảm giá không tồn tại trên hệ thống",
		})
		return
	}

	if !v.IsActive {
		response.Success(c, model.VoucherValidationResult{
			IsValid: false,
			Message: "Mã giảm giá đang tạm ngưng áp dụng",
		})
		return
	}

	now := time.Now()
	if now.Before(v.StartDate) {
		response.Success(c, model.VoucherValidationResult{
			IsValid: false,
			Message: "Mã giảm giá chưa đến ngày bắt đầu áp dụng",
		})
		return
	}
	if now.After(v.EndDate) {
		response.Success(c, model.VoucherValidationResult{
			IsValid: false,
			Message: "Mã giảm giá đã hết hạn sử dụng",
		})
		return
	}
	if v.UsageLimit > 0 && v.UsedCount >= v.UsageLimit {
		response.Success(c, model.VoucherValidationResult{
			IsValid: false,
			Message: "Mã giảm giá đã hết lượt sử dụng",
		})
		return
	}
	if req.SubTotal < v.MinOrderValue {
		response.Success(c, model.VoucherValidationResult{
			IsValid: false,
			Message: "Giá trị đơn hàng chưa đạt mức tối thiểu để áp dụng mã giảm giá này",
		})
		return
	}

	discount := v.CalculateDiscount(req.SubTotal)
	response.Success(c, model.VoucherValidationResult{
		IsValid:        true,
		DiscountAmount: discount,
		Message:        "Áp dụng mã giảm giá thành công",
		Voucher:        v,
	})
}

func (h *VoucherHandler) GetAll(c *gin.Context) {
	hasPagination := c.Query("page") != "" || c.Query("pageSize") != "" || c.Query("limit") != "" || c.Query("search") != ""

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

		filter := repository.VoucherFilter{
			Search: c.Query("search"),
			Limit:  limit,
			Offset: offset,
		}

		list, total, err := h.repo.GetList(c.Request.Context(), filter)
		if err != nil {
			response.HandleError(c, err)
			return
		}
		if list == nil {
			list = []model.Voucher{}
		}

		totalPages := int(math.Ceil(float64(total) / float64(limit)))
		if totalPages == 0 && total == 0 {
			totalPages = 1
		}

		response.SuccessWithMeta(c, list, &response.Meta{
			Page:       int(page),
			PageSize:   int(limit),
			Total:      total,
			TotalPages: totalPages,
		})
		return
	}

	// Cache-Aside: Danh sách toàn bộ voucher (TTL: 10 phút)
	list, err := cache.CacheAside(c.Request.Context(), h.cache, cache.KeyVouchersAll, 10*time.Minute, func() ([]model.Voucher, error) {
		return h.repo.GetAll(c.Request.Context())
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}
	if list == nil {
		list = []model.Voucher{}
	}

	total := int64(len(list))
	response.SuccessWithMeta(c, list, &response.Meta{
		Page:       1,
		PageSize:   int(total),
		Total:      total,
		TotalPages: 1,
	})
}

func (h *VoucherHandler) Create(c *gin.Context) {
	var v model.Voucher
	if err := c.ShouldBindJSON(&v); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	if err := h.repo.Create(c.Request.Context(), &v); err != nil {
		response.HandleError(c, err)
		return
	}

	// Invalidate Cache liên quan đến voucher
	cache.InvalidatePattern(c.Request.Context(), h.cache, cache.PatternVouchers)

	response.CreatedWithMsg(c, "Tạo mã giảm giá mới thành công", v)
}

func (h *VoucherHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, apperror.CodeBadRequest, "Mã voucher không đúng định dạng")
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

	var updated *model.Voucher
	if expectedVersion != nil {
		updated, err = h.repo.Update(c.Request.Context(), id, updates, *expectedVersion)
	} else {
		updated, err = h.repo.Update(c.Request.Context(), id, updates)
	}
	if err != nil {
		response.HandleError(c, err)
		return
	}

	// Invalidate Cache liên quan đến voucher
	cache.InvalidatePattern(c.Request.Context(), h.cache, cache.PatternVouchers)

	response.SuccessWithMsg(c, "Cập nhật mã giảm giá thành công", updated)
}

func (h *VoucherHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, apperror.CodeBadRequest, "Mã voucher không đúng định dạng")
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		response.HandleError(c, err)
		return
	}

	// Invalidate Cache liên quan đến voucher
	cache.InvalidatePattern(c.Request.Context(), h.cache, cache.PatternVouchers)

	response.SuccessWithMsg(c, "Xóa mã giảm giá thành công", nil)
}
