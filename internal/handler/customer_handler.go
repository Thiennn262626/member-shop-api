package handler

import (
	"member-shop-api/internal/model"
	"member-shop-api/internal/repository"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type CustomerHandler struct {
	custRepo repository.CustomerRepository
	txRepo   repository.PointTransactionRepository
}

func NewCustomerHandler(custRepo repository.CustomerRepository, txRepo repository.PointTransactionRepository) *CustomerHandler {
	return &CustomerHandler{custRepo: custRepo, txRepo: txRepo}
}

func (h *CustomerHandler) LookupByPhone(c *gin.Context) {
	phone := c.Query("phone")
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "vui lòng nhập số điện thoại"})
		return
	}

	cust, err := h.custRepo.FindByPhone(c.Request.Context(), phone)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if cust == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "chưa có thông tin hội viên cho số này"})
		return
	}

	txs, _ := h.txRepo.GetByCustomerID(c.Request.Context(), cust.ID)

	c.JSON(http.StatusOK, gin.H{
		"customer":     cust,
		"transactions": txs,
	})
}

func (h *CustomerHandler) GetCustomers(c *gin.Context) {
	list, err := h.custRepo.GetAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func (h *CustomerHandler) AdjustPoints(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mã khách hàng không đúng"})
		return
	}

	var req struct {
		PointsDelta int64  `json:"pointsDelta" binding:"required"`
		Note        string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.custRepo.AdjustPoints(c.Request.Context(), id, req.PointsDelta)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
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

	c.JSON(http.StatusOK, updated)
}
