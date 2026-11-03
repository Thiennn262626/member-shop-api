package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type ShiftStatus string

const (
	ShiftStatusOpen   ShiftStatus = "open"
	ShiftStatusClosed ShiftStatus = "closed"
)

type CashShift struct {
	ID            bson.ObjectID `json:"id" bson:"_id,omitempty"`
	ShiftCode     string        `json:"shiftCode" bson:"shiftCode"` // #CA-MMDD-XX
	StaffID       bson.ObjectID `json:"staffId" bson:"staffId"`
	StaffName     string        `json:"staffName" bson:"staffName"`
	StaffRole     string        `json:"staffRole" bson:"staffRole"`
	ShiftName     string        `json:"shiftName" bson:"shiftName"`
	OpenedAt      time.Time     `json:"openedAt" bson:"openedAt"`
	ClosedAt      *time.Time    `json:"closedAt,omitempty" bson:"closedAt,omitempty"`
	InitialFloat  int64         `json:"initialFloat" bson:"initialFloat"` // Tiền lẻ ban đầu
	CashSales     int64         `json:"cashSales" bson:"cashSales"`       // Tiền mặt từ đơn
	QRSales       int64         `json:"qrSales" bson:"qrSales"`           // Tiền VietQR
	TotalSales    int64         `json:"totalSales" bson:"totalSales"`
	OrderCount    int           `json:"orderCount" bson:"orderCount"`
	ExpectedCash  int64         `json:"expectedCash" bson:"expectedCash"` // initialFloat + cashSales
	ActualCash    *int64        `json:"actualCash,omitempty" bson:"actualCash,omitempty"`
	Difference    *int64        `json:"difference,omitempty" bson:"difference,omitempty"` // actualCash - expectedCash
	Notes         string        `json:"notes,omitempty" bson:"notes,omitempty"`
	NextStaffName string        `json:"nextStaffName,omitempty" bson:"nextStaffName,omitempty"`
	Status        ShiftStatus   `json:"status" bson:"status"`
}
