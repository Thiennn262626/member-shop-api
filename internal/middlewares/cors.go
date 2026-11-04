package middlewares

import (
	"net/http"
	"os"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	// Whitelist các origin được phép kết nối với API
	exactOrigins = map[string]bool{
		"https://member-shop-ui.vercel.app": true,
		"http://localhost:5173":             true,
		"http://127.0.0.1:5173":             true,
		"http://localhost:3000":             true,
	}

	// Regex khớp các preview deployment chính chủ của dự án trên Vercel
	vercelPreviewRegex = regexp.MustCompile(`^https://member-shop(-ui)?(-[a-z0-9]+)?-nguyendinhthiens-projects-8cc2197c\.vercel\.app$`)
)

func isAllowedOrigin(origin string) bool {
	if exactOrigins[origin] {
		return true
	}
	// Cho phép thêm origin qua biến môi trường ALLOWED_ORIGINS (phân cách bằng dấu phẩy)
	customOrigins := os.Getenv("ALLOWED_ORIGINS")
	if customOrigins != "" {
		for _, o := range strings.Split(customOrigins, ",") {
			if strings.TrimSpace(o) == origin {
				return true
			}
		}
	}
	if strings.HasSuffix(origin, ".vercel.app") && vercelPreviewRegex.MatchString(origin) {
		return true
	}
	return false
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		if origin != "" {
			if isAllowedOrigin(origin) {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
			} else {
				// Origin không nằm trong danh sách an toàn -> Từ chối cấp quyền CORS
				if c.Request.Method == http.MethodOptions {
					c.AbortWithStatus(http.StatusForbidden)
					return
				}
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
