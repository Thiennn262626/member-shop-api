package handler

import (
	"member-shop-api/internal/model"
	"member-shop-api/internal/repository"
	"member-shop-api/pkg/apperror"
	"member-shop-api/pkg/response"
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
		response.FailWithStatus(c, http.StatusBadRequest, apperror.CodeBadRequest, "Vui lòng cung cấp số điện thoại khách hàng")
		return
	}

	cust, err := h.custRepo.FindByPhone(c.Request.Context(), phone)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	if cust == nil {
		response.FailWithStatus(c, http.StatusNotFound, apperror.CodeCustomerNotFound, "Chưa có thông tin hội viên cho số điện thoại này")
		return
	}

	txs, _ := h.txRepo.GetByCustomerID(c.Request.Context(), cust.ID)

	response.Success(c, gin.H{
		"customer":     cust,
		"transactions": txs,
	})
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

	response.SuccessWithMsg(c, "Điều chỉnh điểm thưởng khách hàng thành công", updated)
}
