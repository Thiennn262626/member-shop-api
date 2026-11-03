package handler

import (
	"member-shop-api/internal/dto"
	"member-shop-api/internal/model"
	"member-shop-api/internal/repository"
	"member-shop-api/internal/services"
	"member-shop-api/pkg/apperror"
	"member-shop-api/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuthHandler struct {
	authService services.AuthService
	staffRepo   repository.StaffRepository
}

func NewAuthHandler(authService services.AuthService, staffRepo repository.StaffRepository) *AuthHandler {
	return &AuthHandler{authService: authService, staffRepo: staffRepo}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	res, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	response.SuccessWithMsg(c, "Đăng nhập thành công", res)
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userIDStr, _ := c.Get("userId")
	role, _ := c.Get("role")
	username, _ := c.Get("username")
	name, _ := c.Get("name")

	response.Success(c, gin.H{
		"id":       userIDStr,
		"username": username,
		"name":     name,
		"role":     role,
	})
}

func (h *AuthHandler) GetAllStaff(c *gin.Context) {
	list, err := h.staffRepo.GetAll(c.Request.Context())
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.Success(c, list)
}

func (h *AuthHandler) CreateStaff(c *gin.Context) {
	var req struct {
		Username   string `json:"username" binding:"required"`
		Name       string `json:"name" binding:"required"`
		Role       string `json:"role" binding:"required"`
		Phone      string `json:"phone"`
		Shift      string `json:"shift"`
		HourlyRate int64  `json:"hourlyRate"`
		PIN        string `json:"pin" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	staff := &model.StaffAccount{
		Username:   req.Username,
		Name:       req.Name,
		Role:       model.AdminRole(req.Role),
		Phone:      req.Phone,
		Shift:      req.Shift,
		HourlyRate: req.HourlyRate,
		IsActive:   true,
	}
	if err := staff.SetPIN(req.PIN); err != nil {
		response.HandleError(c, apperror.Internal(err, "Không thể mã hóa mã PIN"))
		return
	}

	if err := h.staffRepo.Create(c.Request.Context(), staff); err != nil {
		response.HandleError(c, err)
		return
	}
	response.CreatedWithMsg(c, "Tạo tài khoản nhân viên thành công", staff)
}

func (h *AuthHandler) UpdateStaff(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, apperror.CodeBadRequest, "Mã nhân viên không đúng định dạng")
		return
	}

	var req struct {
		Name       *string `json:"name"`
		Role       *string `json:"role"`
		Phone      *string `json:"phone"`
		Shift      *string `json:"shift"`
		HourlyRate *int64  `json:"hourlyRate"`
		PIN        *string `json:"pin"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	updates := bson.M{}
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Role != nil {
		updates["role"] = *req.Role
	}
	if req.Phone != nil {
		updates["phone"] = *req.Phone
	}
	if req.Shift != nil {
		updates["shift"] = *req.Shift
	}
	if req.HourlyRate != nil {
		updates["hourlyRate"] = *req.HourlyRate
	}
	if req.PIN != nil && *req.PIN != "" {
		temp := &model.StaffAccount{}
		_ = temp.SetPIN(*req.PIN)
		updates["pinHash"] = temp.PINHash
	}

	updated, err := h.staffRepo.Update(c.Request.Context(), id, updates)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMsg(c, "Cập nhật tài khoản nhân viên thành công", updated)
}

func (h *AuthHandler) ToggleStaffStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		response.FailWithStatus(c, http.StatusBadRequest, apperror.CodeBadRequest, "Mã nhân viên không đúng định dạng")
		return
	}

	updated, err := h.staffRepo.ToggleStatus(c.Request.Context(), id)
	if err != nil {
		response.HandleError(c, err)
		return
	}
	response.SuccessWithMsg(c, "Thay đổi trạng thái hoạt động thành công", updated)
}
