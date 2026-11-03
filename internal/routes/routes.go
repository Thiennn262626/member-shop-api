package routes

import (
	"member-shop-api/internal/middlewares"
	"member-shop-api/internal/utils"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

type Route interface {
	Register(v1 *gin.RouterGroup, admin *gin.RouterGroup)
}

func RegisterRoutes(r *gin.Engine, routes ...Route) {
	httpLogger := utils.NewLoggerWithPath("http.log", zerolog.InfoLevel)
	recoveryLogger := utils.NewLoggerWithPath("recovery.log", zerolog.WarnLevel)
	rateLimterLogger := utils.NewLoggerWithPath("rate_limiter.log", zerolog.WarnLevel)

	//dùng gzip để nén kích thước response
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	// Middlewares
	r.Use(
		//middlewares.MetricsMiddleware(),
		middlewares.RateLimiterMiddleware(rateLimterLogger),
		middlewares.CORSMiddleware(),
		middlewares.TraceMiddleware(),
		middlewares.LoggerMiddleware(httpLogger),
		middlewares.RecoveryMiddleware(recoveryLogger),
	)

	r.NoMethod(middlewares.NoMethodHandler())
	r.NoRoute(middlewares.NoRouteHandler())

	v1 := r.Group("/v1")
	admin := r.Group("/admin")

	for _, route := range routes {
		route.Register(v1, admin)
	}
}
