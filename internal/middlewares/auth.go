package middlewares

import (
	"fmt"
	"member-shop-api/internal/config"
	"member-shop-api/pkg/apperror"
	"member-shop-api/pkg/response"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.FailWithStatus(c, http.StatusUnauthorized, apperror.CodeUnauthorized, "Yêu cầu cung cấp Authorization header để xác thực")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			response.FailWithStatus(c, http.StatusUnauthorized, apperror.CodeUnauthorized, "Định dạng token không hợp lệ (phải có tiền tố Bearer)")
			c.Abort()
			return
		}

		tokenStr := parts[1]
		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(cfg.JWTAccessSecret), nil
		})

		if err != nil || !token.Valid {
			response.FailWithStatus(c, http.StatusUnauthorized, apperror.CodeAuthTokenInvalid, "Token xác thực không hợp lệ hoặc đã hết hạn")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			response.FailWithStatus(c, http.StatusUnauthorized, apperror.CodeAuthTokenInvalid, "Claims trong token không hợp lệ")
			c.Abort()
			return
		}

		c.Set("userId", claims["sub"])
		c.Set("role", claims["role"])
		c.Set("username", claims["username"])
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
