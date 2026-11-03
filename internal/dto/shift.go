package dto

type OpenShiftRequest struct {
	ShiftName    string `json:"shiftName" binding:"required"`
	InitialFloat int64  `json:"initialFloat" binding:"min=0"`
	Notes        string `json:"notes,omitempty"`
}

type CloseShiftRequest struct {
	ActualCash    int64  `json:"actualCash" binding:"min=0"`
	Notes         string `json:"notes,omitempty"`
	NextStaffName string `json:"nextStaffName,omitempty"`
}
