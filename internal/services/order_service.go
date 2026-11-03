package services

import (
	"context"
	"fmt"
	"math/rand"
	"member-shop-api/internal/dto"
	"member-shop-api/internal/model"
	"member-shop-api/internal/repository"
	"member-shop-api/pkg/apperror"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

type OrderService interface {
	CreateWebOrder(ctx context.Context, req dto.CreateOrderRequest) (*model.Order, *model.Customer, error)
	CreatePosOrder(ctx context.Context, req dto.CreatePosOrderRequest) (*model.Order, *model.Customer, error)
	UpdateOrderStatus(ctx context.Context, orderID bson.ObjectID, req dto.UpdateOrderStatusRequest) (*model.Order, *model.Customer, error)
}

type orderService struct {
	orderRepo   repository.OrderRepository
	custRepo    repository.CustomerRepository
	productRepo repository.ProductRepository
	voucherRepo repository.VoucherRepository
	shiftRepo   repository.ShiftRepository
	configRepo  repository.ConfigRepository
	pointTxRepo repository.PointTransactionRepository
	sseHub      *SSEHub
}

func NewOrderService(
	orderRepo repository.OrderRepository,
	custRepo repository.CustomerRepository,
	productRepo repository.ProductRepository,
	voucherRepo repository.VoucherRepository,
	shiftRepo repository.ShiftRepository,
	configRepo repository.ConfigRepository,
	pointTxRepo repository.PointTransactionRepository,
	sseHub *SSEHub,
) OrderService {
	return &orderService{
		orderRepo:   orderRepo,
		custRepo:    custRepo,
		productRepo: productRepo,
		voucherRepo: voucherRepo,
		shiftRepo:   shiftRepo,
		configRepo:  configRepo,
		pointTxRepo: pointTxRepo,
		sseHub:      sseHub,
	}
}

func (s *orderService) CreateWebOrder(ctx context.Context, req dto.CreateOrderRequest) (*model.Order, *model.Customer, error) {
	config, err := s.configRepo.GetConfig(ctx)
	if err != nil {
		return nil, nil, err
	}

	customer, err := s.custRepo.CreateOrGet(ctx, req.CustomerName, req.CustomerPhone)
	if err != nil {
		return nil, nil, err
	}

	// 1. Tính toán giá các items an toàn từ backend
	var subTotal int64
	var orderItems []model.OrderItem

	for _, itemReq := range req.Items {
		pID, err := bson.ObjectIDFromHex(itemReq.ProductID)
		if err != nil {
			return nil, nil, fmt.Errorf("mã sản phẩm không hợp lệ: %s", itemReq.ProductID)
		}
		itemTotalPrice := itemReq.UnitPrice * int64(itemReq.Quantity)
		if itemReq.IsRedeemedWithPoints {
			itemTotalPrice = 0 // Đổi món bằng điểm giá = 0đ
		}
		subTotal += itemTotalPrice

		orderItems = append(orderItems, model.OrderItem{
			ProductID:            pID,
			ProductName:          itemReq.ProductName,
			UnitPrice:            itemReq.UnitPrice,
			Quantity:             itemReq.Quantity,
			TotalPrice:           itemTotalPrice,
			OptionsSummary:       itemReq.OptionsSummary,
			ItemNote:             itemReq.ItemNote,
			IsRedeemedWithPoints: itemReq.IsRedeemedWithPoints,
			PointsCost:           itemReq.PointsCost,
		})
	}

	// 2. Tính điểm thưởng trừ tiền mặt
	validPointsToUse := req.DiscountPointsUsed
	if validPointsToUse > customer.AvailablePoints {
		validPointsToUse = customer.AvailablePoints
	}
	discountAmount := validPointsToUse * config.PointToVndRatio

	// 3. Phí vận chuyển
	shippingFee := config.CalculateShippingFee(subTotal, req.FulfillmentType)

	// 4. Áp dụng Voucher
	var voucherDiscount int64
	var appliedCode string
	if req.VoucherCode != "" {
		v, err := s.voucherRepo.GetByCode(ctx, req.VoucherCode)
		if err == nil && v != nil && v.IsActive {
			voucherDiscount = v.CalculateDiscount(subTotal)
			if voucherDiscount > 0 {
				appliedCode = v.Code
				_ = s.voucherRepo.IncrementUsedCount(ctx, v.ID)
			}
		}
	}

	orderTotal := subTotal - voucherDiscount - discountAmount + shippingFee
	if orderTotal < 0 {
		orderTotal = 0
	}

	// Điểm tích lũy dự kiến
	multiplier := model.GetTierMultiplier(customer.Tier)
	earnedPoints := int64(float64(orderTotal/config.SpendToPointRatio) * multiplier)

	orderCode := fmt.Sprintf("#BS-%04d", rand.Intn(9000)+1000)

	newOrder := &model.Order{
		OrderCode:                orderCode,
		CustomerID:               customer.ID,
		CustomerPhone:            customer.Phone,
		CustomerName:             customer.Name,
		CreatedSource:            "web",
		Items:                    orderItems,
		SubTotal:                 subTotal,
		DiscountPointsUsed:       validPointsToUse,
		DiscountAmount:           discountAmount,
		TotalPointsRedeemedItems: req.TotalPointsRedeemedItems,
		FulfillmentType:          req.FulfillmentType,
		DeliveryAddress:          req.DeliveryAddress,
		DeliveryTimeType:         req.DeliveryTimeType,
		ShippingFee:              shippingFee,
		VoucherCode:              appliedCode,
		VoucherDiscount:          voucherDiscount,
		OrderTotal:               orderTotal,
		EarnedPoints:             earnedPoints,
		Note:                     req.Note,
		Status:                   model.OrderStatusPending,
		PaymentStatus:            model.PaymentStatusUnpaid,
		PaymentMethod:            "vietqr",
	}

	if err := s.orderRepo.Create(ctx, newOrder); err != nil {
		return nil, nil, err
	}

	// Broadcast SSE: Đơn hàng mới từ web gửi tới POS quầy
	s.sseHub.Broadcast("order_created", newOrder)

	return newOrder, customer, nil
}

func (s *orderService) CreatePosOrder(ctx context.Context, req dto.CreatePosOrderRequest) (*model.Order, *model.Customer, error) {
	config, err := s.configRepo.GetConfig(ctx)
	if err != nil {
		return nil, nil, err
	}

	customer, err := s.custRepo.CreateOrGet(ctx, req.CustomerName, req.CustomerPhone)
	if err != nil {
		return nil, nil, err
	}

	var subTotal int64
	var orderItems []model.OrderItem

	for _, itemReq := range req.Items {
		pID, _ := bson.ObjectIDFromHex(itemReq.ProductID)
		itemTotalPrice := itemReq.UnitPrice * int64(itemReq.Quantity)
		if itemReq.IsRedeemedWithPoints {
			itemTotalPrice = 0
		}
		subTotal += itemTotalPrice

		orderItems = append(orderItems, model.OrderItem{
			ProductID:            pID,
			ProductName:          itemReq.ProductName,
			UnitPrice:            itemReq.UnitPrice,
			Quantity:             itemReq.Quantity,
			TotalPrice:           itemTotalPrice,
			OptionsSummary:       itemReq.OptionsSummary,
			ItemNote:             itemReq.ItemNote,
			IsRedeemedWithPoints: itemReq.IsRedeemedWithPoints,
			PointsCost:           itemReq.PointsCost,
		})
	}

	validPointsToUse := req.DiscountPointsUsed
	if validPointsToUse > customer.AvailablePoints {
		validPointsToUse = customer.AvailablePoints
	}
	discountAmount := validPointsToUse * config.PointToVndRatio

	var voucherDiscount int64
	var appliedCode string
	if req.VoucherCode != "" {
		v, err := s.voucherRepo.GetByCode(ctx, req.VoucherCode)
		if err == nil && v != nil && v.IsActive {
			voucherDiscount = v.CalculateDiscount(subTotal)
			if voucherDiscount > 0 {
				appliedCode = v.Code
				_ = s.voucherRepo.IncrementUsedCount(ctx, v.ID)
			}
		}
	}

	shippingFee := req.ShippingFee
	if req.FulfillmentType == "delivery" && shippingFee == 0 {
		shippingFee = config.CalculateShippingFee(subTotal, "delivery")
	}

	orderTotal := subTotal - voucherDiscount - discountAmount + shippingFee
	if orderTotal < 0 {
		orderTotal = 0
	}

	multiplier := model.GetTierMultiplier(customer.Tier)
	earnedPoints := int64(float64(orderTotal/config.SpendToPointRatio) * multiplier)

	status := req.Status
	if status == "" {
		status = model.OrderStatusCompleted
	}
	paymentStatus := req.PaymentStatus
	if paymentStatus == "" {
		paymentStatus = model.PaymentStatusPaid
	}

	orderCode := fmt.Sprintf("#POS-%04d", rand.Intn(9000)+1000)
	now := time.Now()

	newOrder := &model.Order{
		OrderCode:                orderCode,
		CustomerID:               customer.ID,
		CustomerPhone:            customer.Phone,
		CustomerName:             customer.Name,
		CreatedSource:            "pos",
		Items:                    orderItems,
		SubTotal:                 subTotal,
		DiscountPointsUsed:       validPointsToUse,
		DiscountAmount:           discountAmount,
		TotalPointsRedeemedItems: req.TotalPointsRedeemedItems,
		FulfillmentType:          req.FulfillmentType,
		DeliveryAddress:          req.DeliveryAddress,
		DeliveryTimeType:         req.DeliveryTimeType,
		ShippingFee:              shippingFee,
		VoucherCode:              appliedCode,
		VoucherDiscount:          voucherDiscount,
		OrderTotal:               orderTotal,
		EarnedPoints:             earnedPoints,
		Note:                     req.Note,
		Status:                   status,
		PaymentStatus:            paymentStatus,
		PaymentMethod:            req.PaymentMethod,
		ShipperInfo:              req.ShipperInfo,
		Timeline: model.OrderTimeline{
			CreatedAt: now,
		},
	}
	if status == model.OrderStatusConfirmed || status == model.OrderStatusCompleted {
		newOrder.Timeline.ConfirmedAt = &now
	}
	if paymentStatus == model.PaymentStatusPaid {
		newOrder.Timeline.PaidAt = &now
	}
	if status == model.OrderStatusCompleted {
		newOrder.Timeline.CompletedAt = &now
	}

	if err := s.orderRepo.Create(ctx, newOrder); err != nil {
		return nil, nil, err
	}

	// Nếu đơn tạo ở trạng thái hoàn tất ngay tại quầy:
	if status == model.OrderStatusCompleted {
		totalPointsDeducted := validPointsToUse + req.TotalPointsRedeemedItems
		updatedCust, err := s.custRepo.UpdateStatsAfterOrder(ctx, customer.ID, orderTotal, earnedPoints, totalPointsDeducted)
		if err == nil {
			customer = updatedCust
		}

		// Ghi nhận sổ giao dịch điểm
		if totalPointsDeducted > 0 {
			_ = s.pointTxRepo.Create(ctx, &model.PointTransaction{
				CustomerID:    customer.ID,
				CustomerPhone: customer.Phone,
				Type:          model.PointTxOrder,
				Source:        model.PointSourceRedeem,
				OrderID:       &newOrder.ID,
				OrderCode:     newOrder.OrderCode,
				Points:        -totalPointsDeducted,
				BalanceAfter:  customer.AvailablePoints - earnedPoints,
				Note:          fmt.Sprintf("Tiêu %d điểm cho đơn POS %s", totalPointsDeducted, newOrder.OrderCode),
			})
		}
		if earnedPoints > 0 {
			_ = s.pointTxRepo.Create(ctx, &model.PointTransaction{
				CustomerID:    customer.ID,
				CustomerPhone: customer.Phone,
				Type:          model.PointTxOrder,
				Source:        model.PointSourceEarn,
				OrderID:       &newOrder.ID,
				OrderCode:     newOrder.OrderCode,
				Points:        earnedPoints,
				BalanceAfter:  customer.AvailablePoints,
				Note:          fmt.Sprintf("Tích điểm hoàn tất đơn POS %s", newOrder.OrderCode),
			})
		}

		// Ghi nhận vào ca trực
		if paymentStatus == model.PaymentStatusPaid {
			_ = s.shiftRepo.RecordSale(ctx, orderTotal, req.PaymentMethod)
		}
	}

	s.sseHub.Broadcast("order_created", newOrder)
	return newOrder, customer, nil
}

func (s *orderService) UpdateOrderStatus(ctx context.Context, orderID bson.ObjectID, req dto.UpdateOrderStatusRequest) (*model.Order, *model.Customer, error) {
	existing, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return nil, nil, apperror.Internal(err, "Không thể truy vấn thông tin đơn hàng")
	}
	if existing == nil {
		return nil, nil, apperror.NotFound("Không tìm thấy đơn hàng yêu cầu cập nhật")
	}

	now := time.Now()
	updates := bson.M{}

	if req.ShipperInfo != nil {
		updates["shipperInfo"] = *req.ShipperInfo
	}
	if req.PaymentStatus != nil {
		updates["paymentStatus"] = *req.PaymentStatus
		if *req.PaymentStatus == model.PaymentStatusPaid {
			updates["timeline.paidAt"] = now
		}
	}

	var statusChangedToCompleted bool
	if req.Status != nil && *req.Status != existing.Status {
		updates["status"] = *req.Status
		switch *req.Status {
		case model.OrderStatusConfirmed:
			updates["timeline.confirmedAt"] = now
		case model.OrderStatusShipping:
			updates["timeline.shippingAt"] = now
		case model.OrderStatusCompleted:
			updates["timeline.completedAt"] = now
			if existing.Status != model.OrderStatusCompleted {
				statusChangedToCompleted = true
			}
		case model.OrderStatusCancelled:
			updates["timeline.cancelledAt"] = now
		}
	}

	updatedOrder, err := s.orderRepo.Update(ctx, orderID, updates)
	if err != nil {
		return nil, nil, err
	}

	customer, _ := s.custRepo.FindByID(ctx, updatedOrder.CustomerID)

	// Nếu đơn vừa chuyển sang hoàn tất: Tích điểm & ghi nhận ca trực
	if statusChangedToCompleted && customer != nil {
		totalDeducted := updatedOrder.DiscountPointsUsed + updatedOrder.TotalPointsRedeemedItems
		updatedCust, err := s.custRepo.UpdateStatsAfterOrder(ctx, customer.ID, updatedOrder.OrderTotal, updatedOrder.EarnedPoints, totalDeducted)
		if err == nil {
			customer = updatedCust
		}

		if totalDeducted > 0 {
			_ = s.pointTxRepo.Create(ctx, &model.PointTransaction{
				CustomerID:    customer.ID,
				CustomerPhone: customer.Phone,
				Type:          model.PointTxOrder,
				Source:        model.PointSourceRedeem,
				OrderID:       &updatedOrder.ID,
				OrderCode:     updatedOrder.OrderCode,
				Points:        -totalDeducted,
				BalanceAfter:  customer.AvailablePoints - updatedOrder.EarnedPoints,
				Note:          fmt.Sprintf("Tiêu %d điểm cho đơn %s", totalDeducted, updatedOrder.OrderCode),
			})
		}
		if updatedOrder.EarnedPoints > 0 {
			_ = s.pointTxRepo.Create(ctx, &model.PointTransaction{
				CustomerID:    customer.ID,
				CustomerPhone: customer.Phone,
				Type:          model.PointTxOrder,
				Source:        model.PointSourceEarn,
				OrderID:       &updatedOrder.ID,
				OrderCode:     updatedOrder.OrderCode,
				Points:        updatedOrder.EarnedPoints,
				BalanceAfter:  customer.AvailablePoints,
				Note:          fmt.Sprintf("Tích điểm hoàn tất đơn %s", updatedOrder.OrderCode),
			})
		}

		if updatedOrder.PaymentStatus == model.PaymentStatusPaid {
			_ = s.shiftRepo.RecordSale(ctx, updatedOrder.OrderTotal, updatedOrder.PaymentMethod)
		}
	}

	s.sseHub.Broadcast("order_updated", updatedOrder)
	return updatedOrder, customer, nil
}
