package routes

import (
	"member-shop-api/internal/config"
	"member-shop-api/internal/handler"
	"member-shop-api/internal/middlewares"

	"github.com/gin-gonic/gin"
)

type WorkShiftRoute struct {
	handler *handler.WorkShiftHandler
	cfg     *config.Config
}

func NewWorkShiftRoute(h *handler.WorkShiftHandler, cfg *config.Config) *WorkShiftRoute {
	return &WorkShiftRoute{handler: h, cfg: cfg}
}

func (r *WorkShiftRoute) Register(v1 *gin.RouterGroup, admin *gin.RouterGroup) {
	shifts := admin.Group("/work-shifts")
	shifts.Use(middlewares.AuthMiddleware(r.cfg))
	{
		shifts.GET("", r.handler.GetWorkShifts)
		shifts.POST("/:id/check-in", r.handler.CheckIn)
		shifts.POST("/roster", middlewares.RequireOwner(), r.handler.SaveRoster)
		shifts.PATCH("/:id/attendance", middlewares.RequireOwner(), r.handler.ConfirmAttendance)
	}

	payroll := admin.Group("/payroll")
	payroll.Use(middlewares.AuthMiddleware(r.cfg), middlewares.RequireOwner())
	{
		payroll.GET("", r.handler.GetPayroll)
		payroll.POST("/adjust/:staffId", r.handler.AdjustPayroll)
		payroll.POST("/finalize", r.handler.FinalizePayroll)
	}
}
