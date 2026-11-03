package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type WorkShiftStatus string

const (
	WorkShiftStatusScheduled      WorkShiftStatus = "scheduled"
	WorkShiftStatusInProgress     WorkShiftStatus = "in_progress"
	WorkShiftStatusCompleted      WorkShiftStatus = "completed"
	WorkShiftStatusLeaveExcused   WorkShiftStatus = "leave_excused"
	WorkShiftStatusLeaveUnexcused WorkShiftStatus = "leave_unexcused"
	WorkShiftStatusSwapped        WorkShiftStatus = "swapped"
)

type WorkShift struct {
	ID                 bson.ObjectID   `json:"id" bson:"_id,omitempty"`
	ShiftCode          string          `json:"shiftCode" bson:"shiftCode"` // #WS-MMDD-XX
	StaffID            bson.ObjectID   `json:"staffId" bson:"staffId"`
	StaffName          string          `json:"staffName" bson:"staffName"`
	StaffRole          string          `json:"staffRole" bson:"staffRole"`
	Date               string          `json:"date" bson:"date"` // YYYY-MM-DD
	ShiftTitle         string          `json:"shiftTitle" bson:"shiftTitle"`
	ShiftType          string          `json:"shiftType" bson:"shiftType"` // 'morning' | 'afternoon'
	ScheduledStartTime string          `json:"scheduledStartTime" bson:"scheduledStartTime"`
	ScheduledEndTime   string          `json:"scheduledEndTime" bson:"scheduledEndTime"`
	ScheduledHours     float64         `json:"scheduledHours" bson:"scheduledHours"`
	ActualStartTime    string          `json:"actualStartTime,omitempty" bson:"actualStartTime,omitempty"`
	ActualEndTime      string          `json:"actualEndTime,omitempty" bson:"actualEndTime,omitempty"`
	ActualHours        *float64        `json:"actualHours,omitempty" bson:"actualHours,omitempty"`
	Status             WorkShiftStatus `json:"status" bson:"status"`
	Notes              string          `json:"notes,omitempty" bson:"notes,omitempty"`
	ActualNotes        string          `json:"actualNotes,omitempty" bson:"actualNotes,omitempty"`
	CreatedAt          time.Time       `json:"createdAt" bson:"createdAt"`
	UpdatedAt          time.Time       `json:"updatedAt" bson:"updatedAt"`
}
