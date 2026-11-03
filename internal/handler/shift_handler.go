package handler

import (
	"fmt"
	"math/rand"
	"member-shop-api/internal/dto"
	"member-shop-api/internal/model"
	"member-shop-api/internal/repository"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ShiftHandler struct {
	repo repository.ShiftRepository
}

func NewShiftHandler(repo repository.ShiftRepository) *ShiftHandler {
	return &ShiftHandler{repo: repo}
}

func (h *ShiftHandler) GetCurrentShift(c *gin.Context) {
	shift, err := h.repo.GetCurrentOpenShift(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if shift == nil {
		c.JSON(http.StatusOK, gin.H{"shift": nil, "message": "không có ca làm việc nào đang mở"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"shift": shift})
}

func (h *ShiftHandler) OpenShift(c *gin.Context) {
	current, _ := h.repo.GetCurrentOpenShift(c.Request.Context())
	if current != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("đang có ca '%s' chưa chốt két, vui lòng chốt két ca cũ trước", current.ShiftName)})
		return
	}

	var req dto.OpenShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr, _ := c.Get("userId")
	name, _ := c.Get("name")
	role, _ := c.Get("role")

	staffID, _ := bson.ObjectIDFromHex(fmt.Sprintf("%v", userIDStr))
	now := time.Now()
	shiftCode := fmt.Sprintf("#CA-%s-%02d", now.Format("0102"), rand.Intn(90)+10)

	newShift := &model.CashShift{
		ShiftCode:    shiftCode,
		StaffID:      staffID,
		StaffName:    fmt.Sprintf("%v", name),
		StaffRole:    fmt.Sprintf("%v", role),
		ShiftName:    req.ShiftName,
		InitialFloat: req.InitialFloat,
		Notes:        req.Notes,
	}

	if err := h.repo.Create(c.Request.Context(), newShift); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, newShift)
}

func (h *ShiftHandler) CloseShift(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mã ca không đúng"})
		return
	}

	var req dto.CloseShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	closed, err := h.repo.Close(c.Request.Context(), id, req.ActualCash, req.Notes, req.NextStaffName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, closed)
}

func (h *ShiftHandler) GetHistory(c *gin.Context) {
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 64)

	history, err := h.repo.GetHistory(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, history)
}
