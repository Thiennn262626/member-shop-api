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

type CustomerHandler struct {
	custRepo repository.CustomerRepository
	txRepo   repository.PointTransactionRepository
	cache    cache.RedisCacheService
}

func NewCustomerHandler(custRepo repository.CustomerRepository, txRepo repository.PointTransactionRepository, cacheSvc cache.RedisCacheService) *CustomerHandler {
	return &CustomerHandler{custRepo: custRepo, txRepo: txRepo, cache: cacheSvc}
}

type CustomerLookupResult struct {
	Customer     *model.Customer          `json:"customer"`
	Transactions []model.PointTransaction `json:"transactions"`
}

func (h *CustomerHandler) LookupByPhone(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		response.FailWithStatus(c, http.StatusBadRequest, apperror.CodeBadRequest, "Vui lòng cung cấp số điện thoại khách hàng")
		return
	}

	// Cache-Aside: Tra cứu hội viên theo số điện thoại (TTL: 3 phút)
	cacheKey := cache.KeyCustomerPhonePrefix + phone
	result, err := cache.CacheAside(c.Request.Context(), h.cache, cacheKey, 3*time.Minute, func() (*CustomerLookupResult, error) {
		cust, err := h.custRepo.FindByPhone(c.Request.Context(), phone)
		if err != nil {
			return nil, err
		}
		if cust == nil {
			return nil, nil
		}

		txs, _ := h.txRepo.GetByCustomerID(c.Request.Context(), cust.ID)
		return &CustomerLookupResult{
			Customer:     cust,
			Transactions: txs,
		}, nil
	})

	if err != nil {
		response.HandleError(c, err)
		return
	}
	if result == nil || result.Customer == nil {
		response.FailWithStatus(c, http.StatusNotFound, apperror.CodeCustomerNotFound, "Chưa có thông tin hội viên cho số điện thoại này")
		return
	}

	response.Success(c, result)
}

func (h *CustomerHandler) GetCustomers(c *gin.Context) {
	list, err := h.custRepo.GetAll(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, list)
}

func (h *CustomerHandler) AdjustPoints(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, apperror.CodeBadRequest, "Mã khách hàng không đúng định dạng")
		return
	}

	var req struct {
		PointsDelta int64  `json:"pointsDelta" binding:"required"`
		Note        string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	updated, err := h.custRepo.AdjustPoints(c.Request.Context(), id, req.PointsDelta)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	// Ghi sổ transaction
	_ = h.txRepo.Create(c.Request.Context(), &model.PointTransaction{
		CustomerID:    updated.ID,
		CustomerPhone: updated.Phone,
		Type:          model.PointTxManual,
		Source:        model.PointSourceAdjust,
		Points:        req.PointsDelta,
		BalanceAfter:  updated.AvailablePoints,
		Note:          req.Note,
	})

	// Invalidate Cache của khách hàng này ngay khi đổi điểm
	cache.Invalidate(c.Request.Context(), h.cache, cache.KeyCustomerPhonePrefix+updated.Phone)

	response.SuccessWithMsg(c, "Điều chỉnh điểm thưởng khách hàng thành công", updated)
}
