package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type MonthlyPayroll struct {
	ID                   bson.ObjectID `json:"id" bson:"_id,omitempty"`
	Month                string        `json:"month" bson:"month"` // YYYY-MM
	StaffID              bson.ObjectID `json:"staffId" bson:"staffId"`
	StaffName            string        `json:"staffName" bson:"staffName"`
	TotalScheduledShifts int           `json:"totalScheduledShifts" bson:"totalScheduledShifts"`
	TotalCompletedShifts int           `json:"totalCompletedShifts" bson:"totalCompletedShifts"`
	TotalLeaveShifts     int           `json:"totalLeaveShifts" bson:"totalLeaveShifts"`
	TotalScheduledHours  float64       `json:"totalScheduledHours" bson:"totalScheduledHours"`
	TotalActualHours     float64       `json:"totalActualHours" bson:"totalActualHours"`
	HourlyRate           int64         `json:"hourlyRate" bson:"hourlyRate"`
	BaseSalary           int64         `json:"baseSalary" bson:"baseSalary"`
	Allowance            int64         `json:"allowance" bson:"allowance"`
	Deduction            int64         `json:"deduction" bson:"deduction"`
	FinalSalary          int64         `json:"finalSalary" bson:"finalSalary"`
	IsFinalized          bool          `json:"isFinalized" bson:"isFinalized"`
	FinalizedAt          *time.Time    `json:"finalizedAt,omitempty" bson:"finalizedAt,omitempty"`
	FinalizedBy          string        `json:"finalizedBy,omitempty" bson:"finalizedBy,omitempty"`
}
