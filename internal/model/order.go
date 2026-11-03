package model

import (
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusShipping  OrderStatus = "shipping"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type PaymentStatus string

const (
	PaymentStatusUnpaid PaymentStatus = "unpaid"
	PaymentStatusPaid   PaymentStatus = "paid"
)

type OrderItem struct {
	ProductID            bson.ObjectID `json:"productId" bson:"productId"`
	ProductName          string        `json:"productName" bson:"productName"`
	UnitPrice            int64         `json:"unitPrice" bson:"unitPrice"`
	Quantity             int           `json:"quantity" bson:"quantity"`
	TotalPrice           int64         `json:"totalPrice" bson:"totalPrice"`
	OptionsSummary       string        `json:"optionsSummary,omitempty" bson:"optionsSummary,omitempty"`
	ItemNote             string        `json:"itemNote,omitempty" bson:"itemNote,omitempty"`
	IsRedeemedWithPoints bool          `json:"isRedeemedWithPoints" bson:"isRedeemedWithPoints"`
	PointsCost           int64         `json:"pointsCost,omitempty" bson:"pointsCost,omitempty"`
}

type OrderTimeline struct {
	CreatedAt   time.Time  `json:"createdAt" bson:"createdAt"`
	ConfirmedAt *time.Time `json:"confirmedAt,omitempty" bson:"confirmedAt,omitempty"`
	PaidAt      *time.Time `json:"paidAt,omitempty" bson:"paidAt,omitempty"`
	ShippingAt  *time.Time `json:"shippingAt,omitempty" bson:"shippingAt,omitempty"`
	CompletedAt *time.Time `json:"completedAt,omitempty" bson:"completedAt,omitempty"`
	CancelledAt *time.Time `json:"cancelledAt,omitempty" bson:"cancelledAt,omitempty"`
}

type Order struct {
	ID                       bson.ObjectID `json:"id" bson:"_id,omitempty"`
	OrderCode                string        `json:"orderCode" bson:"orderCode"` // #BS-xxxx or #POS-xxxx
	CustomerID               bson.ObjectID `json:"customerId" bson:"customerId"`
	CustomerPhone            string        `json:"customerPhone" bson:"customerPhone"`
	CustomerName             string        `json:"customerName" bson:"customerName"`
	CreatedSource            string        `json:"createdSource" bson:"createdSource"` // 'web' | 'pos'
	Items                    []OrderItem   `json:"items" bson:"items"`
	SubTotal                 int64         `json:"subTotal" bson:"subTotal"`
	DiscountPointsUsed       int64         `json:"discountPointsUsed" bson:"discountPointsUsed"`
	DiscountAmount           int64         `json:"discountAmount" bson:"discountAmount"`
	TotalPointsRedeemedItems int64         `json:"totalPointsRedeemedItems" bson:"totalPointsRedeemedItems"`

	FulfillmentType       string     `json:"fulfillmentType" bson:"fulfillmentType"` // 'delivery' | 'pickup'
	DeliveryAddress       string     `json:"deliveryAddress,omitempty" bson:"deliveryAddress,omitempty"`
	DeliveryTimeType      string     `json:"deliveryTimeType" bson:"deliveryTimeType"` // 'asap' | 'scheduled'
	DeliveryScheduledTime *time.Time `json:"deliveryScheduledTime,omitempty" bson:"deliveryScheduledTime,omitempty"`
	ShippingFee           int64      `json:"shippingFee" bson:"shippingFee"`

	VoucherCode     string `json:"voucherCode,omitempty" bson:"voucherCode,omitempty"`
	VoucherDiscount int64  `json:"voucherDiscount" bson:"voucherDiscount"`

	OrderTotal   int64         `json:"orderTotal" bson:"orderTotal"`
	EarnedPoints int64         `json:"earnedPoints" bson:"earnedPoints"`
	Note         string        `json:"note,omitempty" bson:"note,omitempty"`
	Status       OrderStatus   `json:"status" bson:"status"`
	PaymentStatus PaymentStatus `json:"paymentStatus" bson:"paymentStatus"`
	PaymentMethod string        `json:"paymentMethod,omitempty" bson:"paymentMethod,omitempty"` // 'cash' | 'vietqr'
	ShipperInfo   string        `json:"shipperInfo,omitempty" bson:"shipperInfo,omitempty"`

	Timeline OrderTimeline `json:"timeline" bson:"timeline"`
}
