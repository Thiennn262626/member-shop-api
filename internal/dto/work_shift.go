package dto

type RosterAssignmentItem struct {
	Date      string `json:"date" binding:"required"` // YYYY-MM-DD
	ShiftType string `json:"shiftType" binding:"required"` // 'morning' | 'afternoon'
	Notes     string `json:"notes,omitempty"`
}

type SaveRosterRequest struct {
	StaffID     string                 `json:"staffId" binding:"required"`
	Month       string                 `json:"month" binding:"required"` // YYYY-MM
	Assignments []RosterAssignmentItem `json:"assignments" binding:"required"`
}

type ConfirmAttendanceRequest struct {
	ActualHours float64 `json:"actualHours" binding:"required,min=0"`
	Status      string  `json:"status"` // 'completed' | 'leave_excused' | etc.
	Notes       string  `json:"notes,omitempty"`
}

type PayrollAdjustmentRequest struct {
	Allowance int64 `json:"allowance"`
	Deduction int64 `json:"deduction"`
}
