package handler

import (
	"fmt"
	"math/rand"
	"member-shop-api/internal/dto"
	"member-shop-api/internal/model"
	"member-shop-api/internal/repository"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type WorkShiftHandler struct {
	repo      repository.WorkShiftRepository
	staffRepo repository.StaffRepository
}

func NewWorkShiftHandler(repo repository.WorkShiftRepository, staffRepo repository.StaffRepository) *WorkShiftHandler {
	return &WorkShiftHandler{repo: repo, staffRepo: staffRepo}
}

func (h *WorkShiftHandler) GetWorkShifts(c *gin.Context) {
	monthStr := c.Query("month")
	if monthStr == "" {
		monthStr = time.Now().Format("2006-01")
	}

	var staffID *bson.ObjectID
	if sID := c.Query("staffId"); sID != "" {
		if id, err := bson.ObjectIDFromHex(sID); err == nil {
			staffID = &id
		}
	}

	shifts, err := h.repo.GetByMonth(c.Request.Context(), monthStr, staffID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, shifts)
}

func (h *WorkShiftHandler) SaveRoster(c *gin.Context) {
	var req dto.SaveRosterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	staffID, err := bson.ObjectIDFromHex(req.StaffID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mã nhân viên không đúng"})
		return
	}

	staff, err := h.staffRepo.FindByID(c.Request.Context(), staffID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "không tìm thấy nhân viên"})
		return
	}

	// 1. Xóa các ca dự kiến chưa làm trong tháng này của nhân viên này
	_ = h.repo.DeleteFutureScheduled(c.Request.Context(), staffID, req.Month)

	// 2. Tạo mới các ca được xếp
	var newShifts []model.WorkShift
	for i, a := range req.Assignments {
		isMorning := a.ShiftType == "morning"
		start := "06:30"
		end := "14:30"
		title := "Ca Sáng (06:30 - 14:30)"
		if !isMorning {
			start = "14:30"
			end = "22:30"
			title = "Ca Chiều (14:30 - 22:30)"
		}

		dateCode := strings.ReplaceAll(a.Date, "-", "")
		if len(dateCode) >= 4 {
			dateCode = dateCode[4:]
		}

		shiftCode := fmt.Sprintf("#WS-%s-%02d", dateCode, (i+1)+rand.Intn(10))

		newShifts = append(newShifts, model.WorkShift{
			ShiftCode:          shiftCode,
			StaffID:            staff.ID,
			StaffName:          staff.Name,
			StaffRole:          string(staff.Role),
			Date:               a.Date,
			ShiftTitle:         title,
			ShiftType:          a.ShiftType,
			ScheduledStartTime: start,
			ScheduledEndTime:   end,
			ScheduledHours:     8.0,
			Status:             model.WorkShiftStatusScheduled,
			Notes:              a.Notes,
		})
	}

	if err := h.repo.CreateBatch(c.Request.Context(), newShifts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "createdCount": len(newShifts)})
}

func (h *WorkShiftHandler) ConfirmAttendance(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mã ca không đúng"})
		return
	}

	var req dto.ConfirmAttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	status := model.WorkShiftStatusCompleted
	if req.Status != "" {
		status = model.WorkShiftStatus(req.Status)
	}

	updates := bson.M{
		"actualHours": req.ActualHours,
		"status":      status,
		"actualNotes": req.Notes,
	}

	updated, err := h.repo.Update(c.Request.Context(), id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *WorkShiftHandler) CheckIn(c *gin.Context) {
	idStr := c.Param("id")
	id, err := bson.ObjectIDFromHex(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mã ca không đúng"})
		return
	}

	now := time.Now()
	timeStr := fmt.Sprintf("%02d:%02d", now.Hour(), now.Minute())

	updates := bson.M{
		"status":          model.WorkShiftStatusInProgress,
		"actualStartTime": timeStr,
		"actualNotes":     fmt.Sprintf("Đã điểm danh vào ca lúc %s", timeStr),
	}

	updated, err := h.repo.Update(c.Request.Context(), id, updates)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, updated)
}

func (h *WorkShiftHandler) GetPayroll(c *gin.Context) {
	monthStr := c.Query("month")
	if monthStr == "" {
		monthStr = time.Now().Format("2006-01")
	}

	payroll, err := h.repo.GetPayrollSummary(c.Request.Context(), monthStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, payroll)
}

func (h *WorkShiftHandler) AdjustPayroll(c *gin.Context) {
	staffIDStr := c.Param("staffId")
	staffID, err := bson.ObjectIDFromHex(staffIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mã nhân viên không đúng"})
		return
	}

	monthStr := c.Query("month")
	if monthStr == "" {
		monthStr = time.Now().Format("2006-01")
	}

	var req dto.PayrollAdjustmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.repo.SavePayrollAdjustment(c.Request.Context(), monthStr, staffID, req.Allowance, req.Deduction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true})
}

func (h *WorkShiftHandler) FinalizePayroll(c *gin.Context) {
	monthStr := c.Query("month")
	if monthStr == "" {
		monthStr = time.Now().Format("2006-01")
	}

	username, _ := c.Get("username")
	finBy := fmt.Sprintf("%v", username)

	err := h.repo.FinalizePayroll(c.Request.Context(), monthStr, finBy)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "đã chốt bảng lương tháng " + monthStr})
}
