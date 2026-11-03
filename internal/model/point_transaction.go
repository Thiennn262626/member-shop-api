package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type PointTransactionType string

const (
	PointTxOrder  PointTransactionType = "ORDER"
	PointTxReward PointTransactionType = "REWARD"
	PointTxManual PointTransactionType = "MANUAL"
)

type PointTransactionSource string

const (
	PointSourceEarn   PointTransactionSource = "EARN"
	PointSourceRedeem PointTransactionSource = "REDEEM"
	PointSourceAdjust PointTransactionSource = "ADJUST"
)

type PointTransaction struct {
	ID            bson.ObjectID          `json:"id" bson:"_id,omitempty"`
	CustomerID    bson.ObjectID          `json:"customerId" bson:"customerId"`
	CustomerPhone string                 `json:"customerPhone" bson:"customerPhone"`
	Type          PointTransactionType   `json:"type" bson:"type"`
	Source        PointTransactionSource `json:"source" bson:"source"`
	OrderID       *bson.ObjectID         `json:"orderId,omitempty" bson:"orderId,omitempty"`
	OrderCode     string                 `json:"orderCode,omitempty" bson:"orderCode,omitempty"`
	Points        int64                  `json:"points" bson:"points"` // âm khi tiêu, dương khi cộng
	BalanceAfter  int64                  `json:"balanceAfter" bson:"balanceAfter"`
	Note          string                 `json:"note" bson:"note"`
	CreatedAt     time.Time              `json:"createdAt" bson:"createdAt"`
}
