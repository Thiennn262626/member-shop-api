package handler

import (
	"fmt"
	"math/rand"
	"member-shop-api/internal/dto"
	"member-shop-api/internal/model"
	"member-shop-api/internal/repository"
	"member-shop-api/pkg/apperror"
	"member-shop-api/pkg/cache"
	"member-shop-api/pkg/response"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type ShiftHandler struct {
	repo  repository.ShiftRepository
	cache cache.RedisCacheService
}

func NewShiftHandler(repo repository.ShiftRepository, cacheSvc cache.RedisCacheService) *ShiftHandler {
	return &ShiftHandler{repo: repo, cache: cacheSvc}
}

func (h *ShiftHandler) GetCurrentShift(c *gin.Context) {
	// Cache-Aside: Ca trực hiện tại đang mở (TTL: 2 phút)
	shift, err := cache.CacheAside(c.Request.Context(), h.cache, cache.KeyCurrentShift, 2*time.Minute, func() (*model.CashShift, error) {
		return h.repo.GetCurrentOpenShift(c.Request.Context())
	})
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, gin.H{"shift": shift})
}

func (h *ShiftHandler) OpenShift(c *gin.Context) {
	current, _ := h.repo.GetCurrentOpenShift(c.Request.Context())
	if current != nil {
		response.FailWithStatus(c, http.StatusConflict, apperror.CodeShiftAlreadyOpen,
			fmt.Sprintf("Đang có ca '%s' chưa chốt két, vui lòng kết thúc ca cũ trước", current.ShiftName))
		return
	}

	var req dto.OpenShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
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
		response.HandleError(c, err)
		return
	}

	// Invalidate Cache ca trực hiện tại
	cache.Invalidate(c.Request.Context(), h.cache, cache.KeyCurrentShift)

	response.CreatedWithMsg(c, "Mở ca thu ngân thành công", newShift)
}

func (h *ShiftHandler) CloseShift(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, apperror.CodeBadRequest, "Mã ca thu ngân không đúng định dạng")
		return
	}

	var req dto.CloseShiftRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	closed, err := h.repo.Close(c.Request.Context(), id, req.ActualCash, req.Notes, req.NextStaffName)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	// Invalidate Cache ca trực hiện tại
	cache.Invalidate(c.Request.Context(), h.cache, cache.KeyCurrentShift)

	response.SuccessWithMsg(c, "Kết ca và chốt két thành công", closed)
}

func (h *ShiftHandler) GetHistory(c *gin.Context) {
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 64)

	history, err := h.repo.GetHistory(c.Request.Context(), limit, offset)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	if history == nil {
		history = []model.CashShift{}
	}
	response.Success(c, history)
}
