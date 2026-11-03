package middlewares

import (
	"github.com/gin-gonic/gin"
)

func NoMethodHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//response.FailWithDetailed(ctx, http.StatusMethodNotAllowed, nil, "Method Not Allowed")
	}
}

func NoRouteHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		//response.FailWithDetailed(ctx, http.StatusNotFound, nil, "The processing function of the request route was not found")
	}
}
