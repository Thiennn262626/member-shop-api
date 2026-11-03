package handler

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/dto"
	"member-shop-api/internal/model"
	"member-shop-api/internal/repository"
	"member-shop-api/internal/services"
	"member-shop-api/pkg/apperror"
	"member-shop-api/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type AuthHandler struct {
	authService services.AuthService
	staffRepo   repository.StaffRepository
	cfg         *config.Config
}

func NewAuthHandler(authService services.AuthService, staffRepo repository.StaffRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{authService: authService, staffRepo: staffRepo, cfg: cfg}
}

// setRefreshTokenCookie thiết lập cookie HTTP-Only cho Refresh Token bảo mật chống XSS
func (h *AuthHandler) setRefreshTokenCookie(c *gin.Context, token string) {
	secure := h.cfg.AppEnv == "production"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"refresh_token",
		token,
		h.cfg.RefreshTokenExpires,
		"/",
		"",
		secure,
		true, // HttpOnly: JavaScript client không thể đọc, chống rò rỉ XSS
	)
}

// clearRefreshTokenCookie xóa cookie Refresh Token khi đăng xuất
func (h *AuthHandler) clearRefreshTokenCookie(c *gin.Context) {
	secure := h.cfg.AppEnv == "production"
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		"",
		secure,
		true,
	)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.HandleValidationError(c, err)
		return
	}

	res, refreshToken, err := h.authService.Login(c.Request.Context(), req)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	// Thiết lập Refresh Token an toàn trong HTTP-Only Cookie thay vì trả về body
	h.setRefreshTokenCookie(c, refreshToken)

	response.SuccessWithMsg(c, "Đăng nhập thành công", res)
}

func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 1. Đọc Refresh Token ưu tiên từ HTTP-Only Cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil || strings.TrimSpace(refreshToken) == "" {
		// Fallback kiểm tra header X-Refresh-Token hoặc JSON body cho các client API (Postman / mobile client)
		refreshToken = c.GetHeader("X-Refresh-Token")
		if refreshToken == "" {
			var body struct {
				RefreshToken string `json:"refreshToken"`
			}
			_ = c.ShouldBindJSON(&body)
			refreshToken = body.RefreshToken
		}
	}

	if strings.TrimSpace(refreshToken) == "" {
		response.FailWithStatus(c, http.StatusUnauthorized, apperror.CodeUnauthorized, "Yêu cầu cung cấp Refresh Token trong cookie")
		return
	}

	// 2. Thực hiện làm mới token (Token Rotation: Xóa token cũ, thêm vào blacklist, cấp cặp token mới)
	res, newRefreshToken, err := h.authService.RefreshToken(c.Request.Context(), refreshToken)
	if err != nil {
		response.HandleError(c, err)
		return
	}

	// 3. Cập nhật Refresh Token mới vào HTTP-Only Cookie
	h.setRefreshTokenCookie(c, newRefreshToken)

	response.SuccessWithMsg(c, "Làm mới Access Token thành công", res)
}

func (h *AuthHandler) Logout(c *gin.Context) {
	// Lấy Access Token từ header
	authHeader := c.GetHeader("Authorization")
	var accessToken string
	if authHeader != "" {
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			accessToken = parts[1]
		}
	}

	// Lấy Refresh Token từ cookie
	refreshToken, _ := c.Cookie("refresh_token")
	if refreshToken == "" {
		refreshToken = c.GetHeader("X-Refresh-Token")
	}

	// Thu hồi Access Token (đưa jti vào Blacklist) và thu hồi Refresh Token
	_ = h.authService.Logout(c.Request.Context(), accessToken, refreshToken)

	// Xóa cookie Refresh Token trên trình duyệt
	h.clearRefreshTokenCookie(c)

	response.SuccessWithMsg(c, "Đăng xuất thành công", nil)
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
