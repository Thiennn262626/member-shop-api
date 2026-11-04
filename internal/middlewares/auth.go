package middlewares

import (
	"member-shop-api/internal/config"
	"member-shop-api/pkg/apperror"
	"member-shop-api/pkg/authjwt"
	"member-shop-api/pkg/cache"
	"member-shop-api/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	authKM    *authjwt.ECDSAKeyManager
	authCache cache.RedisCacheService
)

// InitAuthMiddleware thiết lập ECDSA Key Manager và Cache Service để kiểm tra chữ ký và blacklist
func InitAuthMiddleware(km *authjwt.ECDSAKeyManager, cacheSvc cache.RedisCacheService) {
	authKM = km
	authCache = cacheSvc
}

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var tokenStr string
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				tokenStr = strings.TrimSpace(parts[1])
			}
		}

		// Hỗ trợ truyền qua query parameter ?token=... chuyên dụng cho EventSource (SSE) / WebSocket
		if tokenStr == "" {
			tokenStr = strings.TrimSpace(c.Query("token"))
		}

		if tokenStr == "" {
			response.FailWithStatus(c, http.StatusUnauthorized, apperror.CodeUnauthorized, "Yêu cầu cung cấp Authorization header hoặc token để xác thực")
			c.Abort()
			return
		}

		// 1. Xác thực bằng chữ ký ECDSA ES256
		if authKM == nil {
			response.FailWithStatus(c, http.StatusInternalServerError, apperror.CodeInternalServerError, "Hệ thống xác thực chưa được khởi tạo")
			c.Abort()
			return
		}

		claims, err := authKM.ValidateAccessToken(tokenStr)
		if err != nil {
			response.FailWithStatus(c, http.StatusUnauthorized, apperror.CodeAuthTokenInvalid, "Token xác thực ECDSA không hợp lệ hoặc đã hết hạn")
			c.Abort()
			return
		}

		// 2. Kiểm tra mã jti của Access Token trong Blacklist (Thu hồi khi Đăng xuất)
		if authCache != nil && claims.JTI != "" {
			jtiKey := cache.KeyBlacklistJTIPrefix + claims.JTI
			isRevoked, _ := authCache.Exists(c.Request.Context(), jtiKey)
			if isRevoked {
				response.FailWithStatus(c, http.StatusUnauthorized, apperror.CodeAuthTokenInvalid, "Token này đã bị thu hồi hoặc đã đăng xuất")
				c.Abort()
				return
			}
		}

		// 3. Gắn thông tin người dùng vào Context cho các Handler phía sau
		c.Set("userId", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("username", claims.Username)
		c.Set("name", claims.Name)
		c.Set("jti", claims.JTI)
		c.Set("accessToken", tokenStr)

		c.Next()
	}
}

func RequireOwner() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "owner" {
			response.FailWithStatus(c, http.StatusForbidden, apperror.CodeForbidden, "Quyền truy cập bị từ chối: Chức năng chỉ dành cho Chủ Cửa Hàng (Owner)")
			c.Abort()
			return
		}
		c.Next()
	}
}
