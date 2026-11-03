package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type VoucherDiscountType string

const (
	VoucherDiscountPercent VoucherDiscountType = "percent"
	VoucherDiscountFixed   VoucherDiscountType = "fixed"
)

type Voucher struct {
	ID                bson.ObjectID       `json:"id" bson:"_id,omitempty"`
	Code              string              `json:"code" bson:"code"`
	Title             string              `json:"title" bson:"title"`
	DiscountType      VoucherDiscountType `json:"discountType" bson:"discountType"`
	DiscountValue     int64               `json:"discountValue" bson:"discountValue"`
	MaxDiscountAmount int64               `json:"maxDiscountAmount,omitempty" bson:"maxDiscountAmount,omitempty"`
	MinOrderValue     int64               `json:"minOrderValue" bson:"minOrderValue"`
	StartDate         time.Time           `json:"startDate" bson:"startDate"`
	EndDate           time.Time           `json:"endDate" bson:"endDate"`
	UsageLimit        int                 `json:"usageLimit,omitempty" bson:"usageLimit,omitempty"`
	UsedCount         int                 `json:"usedCount" bson:"usedCount"`
	IsActive          bool                `json:"isActive" bson:"isActive"`
	CreatedAt         time.Time           `json:"createdAt" bson:"createdAt"`
}

type VoucherValidationResult struct {
	IsValid        bool     `json:"isValid"`
	DiscountAmount int64    `json:"discountAmount"`
	Message        string   `json:"message"`
	Reason         string   `json:"reason,omitempty"`
	Voucher        *Voucher `json:"voucher,omitempty"`
}

// CalculateDiscount calculates the actual discount amount in VND given the subTotal
func (v *Voucher) CalculateDiscount(subTotal int64) int64 {
	if subTotal < v.MinOrderValue {
		return 0
	}
	var discount int64
	if v.DiscountType == VoucherDiscountPercent {
		discount = (subTotal * v.DiscountValue) / 100
		if v.MaxDiscountAmount > 0 && discount > v.MaxDiscountAmount {
			discount = v.MaxDiscountAmount
		}
	} else {
		discount = v.DiscountValue
		if discount > subTotal {
			discount = subTotal
		}
	}
	return discount
}
