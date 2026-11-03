package handler

import (
	"member-shop-api/internal/dto"
	"member-shop-api/internal/model"
	"member-shop-api/internal/repository"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type VoucherHandler struct {
	repo repository.VoucherRepository
}

func NewVoucherHandler(repo repository.VoucherRepository) *VoucherHandler {
	return &VoucherHandler{repo: repo}
}

func (h *VoucherHandler) ValidateVoucher(c *gin.Context) {
	var req dto.ValidateVoucherRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	v, err := h.repo.GetByCode(c.Request.Context(), req.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if v == nil {
		c.JSON(http.StatusOK, model.VoucherValidationResult{
			IsValid: false,
			Message: "mã giảm giá không tồn tại",
		})
		return
	}

	if !v.IsActive {
		c.JSON(http.StatusOK, model.VoucherValidationResult{
			IsValid: false,
			Message: "mã giảm giá đang tạm khóa",
		})
		return
	}

	now := time.Now()
	if now.Before(v.StartDate) {
		c.JSON(http.StatusOK, model.VoucherValidationResult{
			IsValid: false,
			Message: "mã giảm giá chưa đến ngày áp dụng",
		})
		return
	}
	if now.After(v.EndDate) {
		c.JSON(http.StatusOK, model.VoucherValidationResult{
			IsValid: false,
			Message: "mã giảm giá đã hết hạn",
		})
		return
	}
	if v.UsageLimit > 0 && v.UsedCount >= v.UsageLimit {
		c.JSON(http.StatusOK, model.VoucherValidationResult{
			IsValid: false,
			Message: "mã giảm giá đã hết lượt sử dụng",
		})
		return
	}
	if req.SubTotal < v.MinOrderValue {
		c.JSON(http.StatusOK, model.VoucherValidationResult{
			IsValid: false,
			Message: "đơn hàng chưa đạt giá trị tối thiểu để áp dụng mã này",
		})
		return
	}

	discount := v.CalculateDiscount(req.SubTotal)
	c.JSON(http.StatusOK, model.VoucherValidationResult{
		IsValid:        true,
		DiscountAmount: discount,
		Message:        "áp dụng mã giảm giá thành công",
		Voucher:        v,
	})
}

func (h *VoucherHandler) GetAll(c *gin.Context) {
	list, err := h.repo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *VoucherHandler) Create(c *gin.Context) {
	var v model.Voucher
	if err := c.ShouldBindJSON(&v); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.repo.Create(c.Request.Context(), &v); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, v)
}

func (h *VoucherHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mã voucher không đúng"})
		return
	}

	var updates bson.M
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	delete(updates, "_id")
	delete(updates, "id")

	updated, err := h.repo.Update(c.Request.Context(), id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *VoucherHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mã voucher không đúng"})
		return
	}

	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}
