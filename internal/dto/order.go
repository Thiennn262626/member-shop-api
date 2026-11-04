package dto

import "member-shop-api/internal/model"

type OrderItemRequest struct {
	ProductID            string `json:"productId" binding:"required"`
	ProductName          string `json:"productName" binding:"required"`
	UnitPrice            int64  `json:"unitPrice" binding:"required"`
	Quantity             int    `json:"quantity" binding:"required,min=1"`
	OptionsSummary       string `json:"optionsSummary,omitempty"`
	ItemNote             string `json:"itemNote,omitempty"`
	IsRedeemedWithPoints bool   `json:"isRedeemedWithPoints"`
	PointsCost           int64  `json:"pointsCost,omitempty"`
}

type CreateOrderRequest struct {
	CustomerName             string             `json:"customerName" binding:"required"`
	CustomerPhone            string             `json:"customerPhone" binding:"required"`
	Items                    []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
	Note                     string             `json:"note,omitempty"`
	FulfillmentType          string             `json:"fulfillmentType"` // 'delivery' | 'pickup'
	DeliveryAddress          string             `json:"deliveryAddress,omitempty"`
	DeliveryTimeType         string             `json:"deliveryTimeType"` // 'asap' | 'scheduled'
	DeliveryScheduledTime    string             `json:"deliveryScheduledTime,omitempty"`
	VoucherCode              string             `json:"voucherCode,omitempty"`
	DiscountPointsUsed       int64              `json:"discountPointsUsed,omitempty"`
	TotalPointsRedeemedItems int64              `json:"totalPointsRedeemedItems,omitempty"`
}

type CreatePosOrderRequest struct {
	CustomerName             string             `json:"customerName" binding:"required"`
	CustomerPhone            string             `json:"customerPhone" binding:"required"`
	Items                    []OrderItemRequest `json:"items" binding:"required,min=1,dive"`
	Note                     string             `json:"note,omitempty"`
	DiscountPointsUsed       int64              `json:"discountPointsUsed,omitempty"`
	TotalPointsRedeemedItems int64              `json:"totalPointsRedeemedItems,omitempty"`
	PaymentMethod            string             `json:"paymentMethod"` // 'cash' | 'vietqr'
	Status                   model.OrderStatus  `json:"status"`        // 'completed' | 'confirmed'
	PaymentStatus            model.PaymentStatus `json:"paymentStatus"` // 'paid' | 'unpaid'
	ShipperInfo              string             `json:"shipperInfo,omitempty"`
	FulfillmentType          string             `json:"fulfillmentType"`
	DeliveryAddress          string             `json:"deliveryAddress,omitempty"`
	DeliveryTimeType         string             `json:"deliveryTimeType,omitempty"`
	DeliveryScheduledTime    string             `json:"deliveryScheduledTime,omitempty"`
	ShippingFee              int64              `json:"shippingFee,omitempty"`
	VoucherCode              string             `json:"voucherCode,omitempty"`
}

type UpdateOrderStatusRequest struct {
	Status             *model.OrderStatus  `json:"status,omitempty"`
	PaymentStatus      *model.PaymentStatus `json:"paymentStatus,omitempty"`
	ShipperInfo        *string              `json:"shipperInfo,omitempty"`
	DiscountPointsUsed *int64               `json:"discountPointsUsed,omitempty"`
	Version            *int64               `json:"version,omitempty"`
}
