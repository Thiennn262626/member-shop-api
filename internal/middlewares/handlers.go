package middlewares

import (
	"member-shop-api/pkg/response"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NoMethodHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		response.FailWithDetailed(ctx, http.StatusMethodNotAllowed, nil, "Phương thức HTTP không được hỗ trợ trên tuyến đường này")
	}
}

func NoRouteHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		response.FailWithDetailed(ctx, http.StatusNotFound, nil, "Đường dẫn API yêu cầu không tồn tại trên hệ thống")
	}
}
