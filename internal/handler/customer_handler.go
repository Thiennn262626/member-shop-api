package handler

import (
	"member-shop-api/internal/model"
	"member-shop-api/internal/repository"
	"member-shop-api/internal/utils"
	"member-shop-api/pkg/apperror"
	"member-shop-api/pkg/cache"
	"member-shop-api/pkg/response"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CustomerHandler struct {
	custRepo  repository.CustomerRepository
	txRepo    repository.PointTransactionRepository
	orderRepo repository.OrderRepository
	cache     cache.RedisCacheService
}

func NewCustomerHandler(
	custRepo repository.CustomerRepository,
	txRepo repository.PointTransactionRepository,
	orderRepo repository.OrderRepository,
	cacheSvc cache.RedisCacheService,
) *CustomerHandler {
	return &CustomerHandler{
		custRepo:  custRepo,
		txRepo:    txRepo,
		orderRepo: orderRepo,
		cache:     cacheSvc,
	}
}

type CustomerLookupResult struct {
	Customer     *model.Customer          `json:"customer"`
	Orders       []model.Order            `json:"orders"`
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

		orders, _ := h.orderRepo.GetByPhone(c.Request.Context(), cust.Phone)
		if orders == nil {
			orders = []model.Order{}
		}
		txs, _ := h.txRepo.GetByCustomerID(c.Request.Context(), cust.ID)
		if txs == nil {
			txs = []model.PointTransaction{}
		}
		return &CustomerLookupResult{
			Customer:     cust,
			Orders:       orders,
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

	// Bảo vệ dữ liệu cá nhân (SEC-04): Mask PII trên endpoint public chống crawler/scrapers
	maskedCust := *result.Customer
	maskedCust.Name = utils.MaskName(result.Customer.Name)
	maskedCust.Phone = utils.MaskPhone(result.Customer.Phone)

	maskedOrders := make([]model.Order, len(result.Orders))
	for i, o := range result.Orders {
		maskedO := o
		maskedO.CustomerName = utils.MaskName(o.CustomerName)
		maskedO.CustomerPhone = utils.MaskPhone(o.CustomerPhone)
		maskedO.DeliveryAddress = utils.MaskAddress(o.DeliveryAddress)
		maskedOrders[i] = maskedO
	}

	maskedTxs := make([]model.PointTransaction, len(result.Transactions))
	for i, tx := range result.Transactions {
		maskedTx := tx
		maskedTx.CustomerPhone = utils.MaskPhone(tx.CustomerPhone)
		maskedTxs[i] = maskedTx
	}

	response.Success(c, &CustomerLookupResult{
		Customer:     &maskedCust,
		Orders:       maskedOrders,
		Transactions: maskedTxs,
	})
}

func (h *CustomerHandler) GetCustomers(c *gin.Context) {
	list, err := h.custRepo.GetAll(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	if list == nil {
		list = []model.Customer{}
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
