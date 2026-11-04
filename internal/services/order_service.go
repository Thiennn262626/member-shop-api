package services

import (
	"context"
	"fmt"
	"math/rand"
	"member-shop-api/internal/dto"
	"member-shop-api/internal/model"
	"member-shop-api/internal/repository"
	"member-shop-api/pkg/apperror"
	"member-shop-api/pkg/cache"
	"net/http"
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
	cache       cache.RedisCacheService
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
	cacheSvc cache.RedisCacheService,
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
		cache:       cacheSvc,
	}
}

func (s *orderService) CreateWebOrder(ctx context.Context, req dto.CreateOrderRequest) (*model.Order, *model.Customer, error) {
	// Cache-Aside: Truy vấn cấu hình từ Cache (1h)
	config, err := cache.CacheAside(ctx, s.cache, cache.KeyShopConfig, 1*time.Hour, func() (*model.ShopConfig, error) {
		return s.configRepo.GetConfig(ctx)
	})
	if err != nil {
		return nil, nil, err
	}

	customer, err := s.custRepo.CreateOrGet(ctx, req.CustomerName, req.CustomerPhone)
	if err != nil {
		return nil, nil, err
	}

	// 1. Tính toán giá các items an toàn từ backend (Bảo vệ chống Client Price Tampering)
	var subTotal int64
	var orderItems []model.OrderItem
	var calculatedRedeemPoints int64

	for _, itemReq := range req.Items {
		pID, err := bson.ObjectIDFromHex(itemReq.ProductID)
		if err != nil {
			return nil, nil, fmt.Errorf("mã sản phẩm không hợp lệ: %s", itemReq.ProductID)
		}

		// Xác thực trực tiếp giá gốc của sản phẩm từ Database
		prod, err := s.productRepo.GetByID(ctx, pID)
		if err != nil || prod == nil {
			return nil, nil, fmt.Errorf("không tìm thấy thông tin sản phẩm: %s", itemReq.ProductName)
		}
		if !prod.IsActive {
			return nil, nil, fmt.Errorf("sản phẩm '%s' hiện đang tạm ngưng mở bán", prod.Name)
		}

		// Giá cơ sở bắt buộc từ DB
		realUnitPrice := prod.Price
		if itemReq.UnitPrice > prod.Price {
			toppingDelta := itemReq.UnitPrice - prod.Price
			if toppingDelta > 0 && toppingDelta < 100000 {
				realUnitPrice += toppingDelta
			}
		}

		itemTotalPrice := realUnitPrice * int64(itemReq.Quantity)
		var pointsCost int64

		if itemReq.IsRedeemedWithPoints {
			if !prod.AllowPointRedemption {
				return nil, nil, fmt.Errorf("món '%s' không hỗ trợ đổi bằng điểm thưởng", prod.Name)
			}
			ptsRequired := prod.PointsRequiredToRedeem
			if ptsRequired <= 0 {
				ptsRequired = prod.Price / config.PointToVndRatio
				if ptsRequired <= 0 {
					ptsRequired = 1
				}
			}
			pointsCost = ptsRequired * int64(itemReq.Quantity)
			calculatedRedeemPoints += pointsCost
			itemTotalPrice = 0 // Đổi món bằng điểm giá = 0đ
		}

		subTotal += itemTotalPrice

		orderItems = append(orderItems, model.OrderItem{
			ProductID:            pID,
			ProductName:          prod.Name,
			UnitPrice:            realUnitPrice,
			Quantity:             itemReq.Quantity,
			TotalPrice:           itemTotalPrice,
			OptionsSummary:       itemReq.OptionsSummary,
			ItemNote:             itemReq.ItemNote,
			IsRedeemedWithPoints: itemReq.IsRedeemedWithPoints,
			PointsCost:           pointsCost,
		})
	}

	// Xác thực số điểm đổi món có vượt quá điểm khả dụng của khách không
	if calculatedRedeemPoints > customer.AvailablePoints {
		return nil, nil, apperror.New(http.StatusBadRequest, apperror.CodeInsufficientPoints, fmt.Sprintf("Bạn cần %d điểm để đổi các món này nhưng chỉ có %d điểm khả dụng", calculatedRedeemPoints, customer.AvailablePoints))
	}

	// 2. Tính điểm thưởng trừ tiền mặt (sau khi trừ đi điểm đã dùng để đổi món)
	remainingPoints := customer.AvailablePoints - calculatedRedeemPoints
	validPointsToUse := req.DiscountPointsUsed
	if validPointsToUse > remainingPoints {
		validPointsToUse = remainingPoints
	}
	discountAmount := validPointsToUse * config.PointToVndRatio

	// Chống Double-Spending (SEC-06): Khóa/trừ điểm Atomic ngay lập tức khi tạo đơn
	totalPointsToDeduct := validPointsToUse + calculatedRedeemPoints
	if totalPointsToDeduct > 0 {
		updatedCust, err := s.custRepo.DeductPointsAtomic(ctx, customer.ID, totalPointsToDeduct)
		if err != nil {
			return nil, nil, apperror.New(http.StatusBadRequest, apperror.CodeInsufficientPoints, "Không đủ điểm thưởng khả dụng hoặc điểm đã được sử dụng ở đơn hàng khác")
		}
		customer = updatedCust
	}

	// 3. Phí vận chuyển
	shippingFee := config.CalculateShippingFee(subTotal, req.FulfillmentType)

	// 4. Áp dụng Voucher (Cache-Aside)
	var voucherDiscount int64
	var appliedCode string
	if req.VoucherCode != "" {
		vKey := cache.KeyVoucherPrefix + req.VoucherCode
		v, err := cache.CacheAside(ctx, s.cache, vKey, 10*time.Minute, func() (*model.Voucher, error) {
			return s.voucherRepo.GetByCode(ctx, req.VoucherCode)
		})
		if err == nil && v != nil && v.IsActive {
			voucherDiscount = v.CalculateDiscount(subTotal)
			if voucherDiscount > 0 {
				appliedCode = v.Code
				_ = s.voucherRepo.IncrementUsedCount(ctx, v.ID)
				cache.Invalidate(ctx, s.cache, vKey)
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
		// Rollback hoàn điểm nếu tạo đơn hàng thất bại
		if totalPointsToDeduct > 0 {
			_, _ = s.custRepo.RefundPointsAtomic(ctx, customer.ID, totalPointsToDeduct)
		}
		return nil, nil, err
	}

	// Ghi nhận sổ giao dịch điểm trừ thành công
	if totalPointsToDeduct > 0 {
		_ = s.pointTxRepo.Create(ctx, &model.PointTransaction{
			CustomerID:    customer.ID,
			CustomerPhone: customer.Phone,
			Type:          model.PointTxOrder,
			Source:        model.PointSourceRedeem,
			OrderID:       &newOrder.ID,
			OrderCode:     newOrder.OrderCode,
			Points:        -totalPointsToDeduct,
			BalanceAfter:  customer.AvailablePoints,
			Note:          fmt.Sprintf("Sử dụng %d điểm khi đặt đơn %s", totalPointsToDeduct, newOrder.OrderCode),
		})
	}

	// Invalidate Cache của khách hàng để cập nhật thông tin điểm mới nhất
	cache.Invalidate(ctx, s.cache, cache.KeyCustomerPhonePrefix+customer.Phone)

	// Broadcast SSE: Đơn hàng mới từ web gửi tới POS quầy
	s.sseHub.Broadcast("order_created", newOrder)

	return newOrder, customer, nil
}

func (s *orderService) CreatePosOrder(ctx context.Context, req dto.CreatePosOrderRequest) (*model.Order, *model.Customer, error) {
	// Cache-Aside: Truy vấn cấu hình từ Cache (1h)
	config, err := cache.CacheAside(ctx, s.cache, cache.KeyShopConfig, 1*time.Hour, func() (*model.ShopConfig, error) {
		return s.configRepo.GetConfig(ctx)
	})
	if err != nil {
		return nil, nil, err
	}

	customer, err := s.custRepo.CreateOrGet(ctx, req.CustomerName, req.CustomerPhone)
	if err != nil {
		return nil, nil, err
	}

	var subTotal int64
	var orderItems []model.OrderItem
	var calculatedRedeemPoints int64

	for _, itemReq := range req.Items {
		pID, _ := bson.ObjectIDFromHex(itemReq.ProductID)
		prod, err := s.productRepo.GetByID(ctx, pID)
		realUnitPrice := itemReq.UnitPrice
		if err == nil && prod != nil {
			realUnitPrice = prod.Price
			if itemReq.UnitPrice > prod.Price {
				toppingDelta := itemReq.UnitPrice - prod.Price
				if toppingDelta > 0 && toppingDelta < 100000 {
					realUnitPrice += toppingDelta
				}
			}
		}

		itemTotalPrice := realUnitPrice * int64(itemReq.Quantity)
		var pointsCost int64
		if itemReq.IsRedeemedWithPoints {
			if prod != nil && prod.AllowPointRedemption {
				ptsRequired := prod.PointsRequiredToRedeem
				if ptsRequired <= 0 {
					ptsRequired = prod.Price / config.PointToVndRatio
					if ptsRequired <= 0 {
						ptsRequired = 1
					}
				}
				pointsCost = ptsRequired * int64(itemReq.Quantity)
				calculatedRedeemPoints += pointsCost
			}
			itemTotalPrice = 0
		}
		subTotal += itemTotalPrice

		orderItems = append(orderItems, model.OrderItem{
			ProductID:            pID,
			ProductName:          itemReq.ProductName,
			UnitPrice:            realUnitPrice,
			Quantity:             itemReq.Quantity,
			TotalPrice:           itemTotalPrice,
			OptionsSummary:       itemReq.OptionsSummary,
			ItemNote:             itemReq.ItemNote,
			IsRedeemedWithPoints: itemReq.IsRedeemedWithPoints,
			PointsCost:           pointsCost,
		})
	}

	remainingPoints := customer.AvailablePoints - calculatedRedeemPoints
	if remainingPoints < 0 {
		remainingPoints = 0
	}
	validPointsToUse := req.DiscountPointsUsed
	if validPointsToUse > remainingPoints {
		validPointsToUse = remainingPoints
	}
	discountAmount := validPointsToUse * config.PointToVndRatio

	// Chống Double-Spending (SEC-06): Khóa/trừ điểm Atomic ngay lập tức khi tạo đơn POS
	totalPointsDeducted := validPointsToUse + calculatedRedeemPoints
	if totalPointsDeducted > 0 {
		updatedCust, err := s.custRepo.DeductPointsAtomic(ctx, customer.ID, totalPointsDeducted)
		if err != nil {
			return nil, nil, apperror.New(http.StatusBadRequest, apperror.CodeInsufficientPoints, "Không đủ điểm thưởng khả dụng hoặc điểm đã được sử dụng ở đơn hàng khác")
		}
		customer = updatedCust
	}

	var voucherDiscount int64
	var appliedCode string
	if req.VoucherCode != "" {
		vKey := cache.KeyVoucherPrefix + req.VoucherCode
		v, err := cache.CacheAside(ctx, s.cache, vKey, 10*time.Minute, func() (*model.Voucher, error) {
			return s.voucherRepo.GetByCode(ctx, req.VoucherCode)
		})
		if err == nil && v != nil && v.IsActive {
			voucherDiscount = v.CalculateDiscount(subTotal)
			if voucherDiscount > 0 {
				appliedCode = v.Code
				_ = s.voucherRepo.IncrementUsedCount(ctx, v.ID)
				cache.Invalidate(ctx, s.cache, vKey)
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
		CreatedAt:                now,
	}
	if status == model.OrderStatusConfirmed || status == model.OrderStatusCompleted {
		newOrder.ConfirmedAt = &now
	}
	if paymentStatus == model.PaymentStatusPaid {
		newOrder.PaidAt = &now
	}
	if status == model.OrderStatusCompleted {
		newOrder.CompletedAt = &now
	}

	if err := s.orderRepo.Create(ctx, newOrder); err != nil {
		if totalPointsDeducted > 0 {
			_, _ = s.custRepo.RefundPointsAtomic(ctx, customer.ID, totalPointsDeducted)
		}
		return nil, nil, err
	}

	// Ghi nhận sổ giao dịch điểm tạm trừ cho đơn POS
	if totalPointsDeducted > 0 {
		_ = s.pointTxRepo.Create(ctx, &model.PointTransaction{
			CustomerID:    customer.ID,
			CustomerPhone: customer.Phone,
			Type:          model.PointTxOrder,
			Source:        model.PointSourceRedeem,
			OrderID:       &newOrder.ID,
			OrderCode:     newOrder.OrderCode,
			Points:        -totalPointsDeducted,
			BalanceAfter:  customer.AvailablePoints,
			Note:          fmt.Sprintf("Sử dụng %d điểm cho đơn POS %s", totalPointsDeducted, newOrder.OrderCode),
		})
	}

	// Nếu đơn tạo ở trạng thái hoàn tất ngay tại quầy:
	if status == model.OrderStatusCompleted {
		updatedCust, err := s.custRepo.AddSpentAndEarnedPoints(ctx, customer.ID, orderTotal, earnedPoints)
		if err == nil {
			customer = updatedCust
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

	// Invalidate Cache liên quan
	cache.Invalidate(ctx, s.cache, cache.KeyCustomerPhonePrefix+customer.Phone, cache.KeyCurrentShift)

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
			updates["paidAt"] = now
		}
	}

	var statusChangedToCompleted bool
	var statusChangedToCancelled bool
	if req.Status != nil && *req.Status != existing.Status {
		updates["status"] = *req.Status
		switch *req.Status {
		case model.OrderStatusConfirmed:
			updates["confirmedAt"] = now
		case model.OrderStatusShipping:
			updates["shippingAt"] = now
		case model.OrderStatusCompleted:
			updates["completedAt"] = now
			if existing.Status != model.OrderStatusCompleted {
				statusChangedToCompleted = true
			}
		case model.OrderStatusCancelled:
			updates["cancelledAt"] = now
			if existing.Status != model.OrderStatusCancelled {
				statusChangedToCancelled = true
			}
		}
	}

	// Xử lý hoàn điểm nếu đơn bị hủy (SEC-06)
	customer, _ := s.custRepo.FindByID(ctx, existing.CustomerID)
	if statusChangedToCancelled && customer != nil {
		pointsToRefund := existing.DiscountPointsUsed + existing.TotalPointsRedeemedItems
		if pointsToRefund > 0 && !existing.PointsRefunded {
			updatedCust, err := s.custRepo.RefundPointsAtomic(ctx, customer.ID, pointsToRefund)
			if err == nil {
				customer = updatedCust
				updates["pointsRefunded"] = true

				_ = s.pointTxRepo.Create(ctx, &model.PointTransaction{
					CustomerID:    customer.ID,
					CustomerPhone: customer.Phone,
					Type:          model.PointTxOrder,
					Source:        model.PointSourceAdjust,
					OrderID:       &existing.ID,
					OrderCode:     existing.OrderCode,
					Points:        pointsToRefund,
					BalanceAfter:  customer.AvailablePoints,
					Note:          fmt.Sprintf("Hoàn lại %d điểm do hủy đơn %s", pointsToRefund, existing.OrderCode),
				})
			}
		}
	}

	updatedOrder, err := s.orderRepo.Update(ctx, orderID, updates)
	if err != nil {
		return nil, nil, err
	}

	// Nếu đơn vừa chuyển sang hoàn tất: Tích điểm & ghi nhận ca trực
	if statusChangedToCompleted && customer != nil {
		updatedCust, err := s.custRepo.AddSpentAndEarnedPoints(ctx, customer.ID, updatedOrder.OrderTotal, updatedOrder.EarnedPoints)
		if err == nil {
			customer = updatedCust
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

	// Invalidate Cache liên quan khi trạng thái đơn thay đổi
	cache.Invalidate(ctx, s.cache, cache.KeyCustomerPhonePrefix+existing.CustomerPhone, cache.KeyCurrentShift, "shop:order:code:"+existing.OrderCode)

	s.sseHub.Broadcast("order_updated", updatedOrder)
	return updatedOrder, customer, nil
}
