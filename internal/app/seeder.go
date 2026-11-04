package app

import (
	"context"
	"member-shop-api/internal/model"
	"member-shop-api/pkg/logger"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
)

func SeedInitialData(mc *ModuleContext) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Seed & Sync Products: Đảm bảo danh mục chỉ lưu chuẩn xác 4 món Bánh Tráng đặc trưng (giá 15k - 25k)
	existingProducts, err := mc.ProductRepo.GetAll(ctx, false)
	needsSync := false
	if err != nil || len(existingProducts) != 4 {
		needsSync = true
	} else {
		for _, p := range existingProducts {
			if p.Category != "Bánh Tráng Trộn" && p.Category != "Bánh Tráng Cuốn" && p.Category != "Bánh Tráng Chấm" && p.Category != "Bánh Tráng Nướng" {
				needsSync = true
				break
			}
		}
	}

	if needsSync {
		// Xóa sạch collection sản phẩm cũ để đồng bộ chuẩn 4 món Bánh Tráng
		_, _ = mc.DB.Collection("products").DeleteMany(ctx, bson.M{})

		now := time.Now()

		spiceOptions := model.ItemOptionGroup{
			ID:        "spice",
			Name:      "Mức Độ Cay",
			Required:  true,
			Selection: "only",
			Options: []model.ItemOptionItem{
				{ID: "sp-mild", Name: "Cay vừa (Tiêu chuẩn thơm ngon)", Price: 0, IsDefault: true, IsAvailable: true},
				{ID: "sp-none", Name: "Không cay (Không bỏ ớt sa tế)", Price: 0, IsDefault: false, IsAvailable: true},
				{ID: "sp-hot", Name: "Siêu cay (Hít hà tê tái)", Price: 0, IsDefault: false, IsAvailable: true},
			},
		}

		customNoteOptions := model.ItemOptionGroup{
			ID:        "custom",
			Name:      "Lưu Ý Khẩu Vị",
			Required:  false,
			Selection: "many",
			Options: []model.ItemOptionItem{
				{ID: "no-coriander", Name: "Không lấy rau răm", Price: 0, IsAvailable: true},
				{ID: "no-onion", Name: "Không lấy hành phi", Price: 0, IsAvailable: true},
				{ID: "sauce-separate", Name: "Để nước sốt riêng đi xa", Price: 0, IsAvailable: true},
			},
		}

		toppingOptions := model.ItemOptionGroup{
			ID:        "toppings",
			Name:      "Topping Thêm (Ăn Cực Đã)",
			Required:  false,
			Selection: "many",
			Options: []model.ItemOptionItem{
				{ID: "top-quail", Name: "Thêm 3 Trứng Cút Lòng Đào", Price: 5000, IsAvailable: true},
				{ID: "top-beef", Name: "Thêm Khô Bò Đen Xé Sợi", Price: 5000, IsAvailable: true},
				{ID: "top-chicken", Name: "Thêm Khô Gà Lá Chanh", Price: 5000, IsAvailable: true},
				{ID: "top-butter", Name: "Thêm Bịch Sốt Bơ Béo", Price: 5000, IsAvailable: true},
				{ID: "top-mango", Name: "Thêm Xoài Chua Bào Sợi", Price: 3000, IsAvailable: true},
				{ID: "top-shrimp", Name: "Thêm Tép Sấy Giòn Tan", Price: 3000, IsAvailable: true},
			},
		}

		products := []model.Product{
			{
				Name:                 "Bánh Tráng Trộn Thập Cẩm Tây Ninh",
				Category:             "Bánh Tráng Trộn",
				Description:          "Bánh tráng phơi sương dẻo thơm thấm muối tôm, khô bò đen xé sợi, tép sấy giòn tan, trứng cút lòng đào, xoài bào chua và hành phi thơm lừng.",
				Price:                20000,
				FakePrice:            25000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1546069901-ba9599a7e63c?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Bịch",
				Tag:                  "Bán Chạy 🔥",
				Rating:               5.0,
				SoldCount:            1450,
				OptionGroups:         []model.ItemOptionGroup{spiceOptions, customNoteOptions, toppingOptions},
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Bánh Tráng Cuốn Sốt Bơ Trứng Cút",
				Category:             "Bánh Tráng Cuốn",
				Description:          "Bánh tráng cuốn dẻo ngập tràn khô bò, tép sấy, trứng cút lòng đào, chấm đẫm sốt bơ béo ngậy vàng óng kèm sốt me chua ngọt đậm đà.",
				Price:                25000,
				FakePrice:            30000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1540420773420-3366772f4999?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Hộp 5 cuốn",
				Tag:                  "Signature ⭐",
				Rating:               4.9,
				SoldCount:            1120,
				OptionGroups:         []model.ItemOptionGroup{spiceOptions, customNoteOptions, toppingOptions},
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Bánh Tráng Chấm Sốt Tắc Me Rim Cay",
				Category:             "Bánh Tráng Chấm",
				Description:          "Bánh tráng phơi sương dẻo quẹt sốt me tắc cay chua sánh đặc, rắc đẫm hành phi giòn tan và muối tôm hạt to ăn bao ghiền.",
				Price:                15000,
				FakePrice:            18000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1555126634-323283e090fa?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Phần",
				Tag:                  "Gây Nghiện 🌶️",
				Rating:               4.8,
				SoldCount:            890,
				OptionGroups:         []model.ItemOptionGroup{spiceOptions, toppingOptions},
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Bánh Tráng Nướng Mỡ Hành Trứng Cút",
				Category:             "Bánh Tráng Nướng",
				Description:          "Bánh tráng nướng than hoa nóng hổi giòn rụm, quét mỡ hành thơm phức, trứng cút béo ngậy, xúc xích và tép sấy giòn tan.",
				Price:                20000,
				FakePrice:            25000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1565299585323-38d6b0865b47?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Cái",
				Tag:                  "Nóng Giòn ♨️",
				Rating:               4.9,
				SoldCount:            780,
				OptionGroups:         []model.ItemOptionGroup{spiceOptions},
				CreatedAt:            now,
				UpdatedAt:            now,
			},
		}

		for _, p := range products {
			_ = mc.ProductRepo.Create(ctx, &p)
		}

		// Xóa cache Redis để đồng bộ ngay lập tức cho API storefront
		if mc.Redis != nil {
			_ = mc.Redis.Del(ctx, "shop:products:active")
			_ = mc.Redis.Del(ctx, "shop:products:all")
		}

		if logger.Log != nil {
			logger.Log.Info().Int("count", len(products)).Msg("🌶️ Synced exactly 4 signature Bánh Tráng products to MongoDB Atlas & cleared Redis cache")
		}
	}

	// 2. Seed Vouchers if empty
	existingVouchers, err := mc.VoucherRepo.GetAll(ctx)
	if err == nil && len(existingVouchers) == 0 {
		startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
		vouchers := []model.Voucher{
			{
				Code:          "BANHTRANG5K",
				Title:         "Giảm 5.000đ Cho Đơn Từ 40k",
				DiscountType:  model.VoucherDiscountFixed,
				DiscountValue: 5000,
				MinOrderValue: 40000,
				StartDate:     startDate,
				EndDate:       endDate,
				UsageLimit:    500,
				UsedCount:     0,
				IsActive:      true,
				CreatedAt:     time.Now(),
			},
			{
				Code:              "ANVAT10",
				Title:             "Giảm 10% Tối Đa 15k (Đơn từ 50k)",
				DiscountType:      model.VoucherDiscountPercent,
				DiscountValue:     10,
				MaxDiscountAmount: 15000,
				MinOrderValue:     50000,
				StartDate:         startDate,
				EndDate:           endDate,
				UsageLimit:        500,
				UsedCount:         0,
				IsActive:          true,
				CreatedAt:         time.Now(),
			},
			{
				Code:          "FREESHIP",
				Title:         "Hỗ Trợ 15k Phí Ship (Đơn từ 70k)",
				DiscountType:  model.VoucherDiscountFixed,
				DiscountValue: 15000,
				MinOrderValue: 70000,
				StartDate:     startDate,
				EndDate:       endDate,
				UsageLimit:    300,
				UsedCount:     0,
				IsActive:      true,
				CreatedAt:     time.Now(),
			},
		}
		for _, v := range vouchers {
			_ = mc.VoucherRepo.Create(ctx, &v)
		}
		if logger.Log != nil {
			logger.Log.Info().Int("count", len(vouchers)).Msg("🎟️ Seeded initial vouchers to MongoDB")
		}
	}

	// 3. Seed Customers if empty
	existingCustomers, err := mc.CustRepo.GetAll(ctx)
	if err == nil && len(existingCustomers) == 0 {
		customers := []model.Customer{
			{
				Phone:           "0988111222",
				Name:            "Linh Chi (Chi Chi)",
				Tier:            model.TierDiamond,
				EarnedPoints:    180,
				RedeemedPoints:  30,
				AvailablePoints: 150,
				TotalSpent:      1650000,
				JoinedDate:      "15/12/2025",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			{
				Phone:           "0977333444",
				Name:            "Minh Quân (Quân Bùi)",
				Tier:            model.TierGold,
				EarnedPoints:    85,
				RedeemedPoints:  15,
				AvailablePoints: 70,
				TotalSpent:      720000,
				JoinedDate:      "10/01/2026",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			{
				Phone:           "0912345678",
				Name:            "Phương Thảo",
				Tier:            model.TierSilver,
				EarnedPoints:    30,
				RedeemedPoints:  10,
				AvailablePoints: 20,
				TotalSpent:      240000,
				JoinedDate:      "01/03/2026",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
		}
		for _, c := range customers {
			cust, err := mc.CustRepo.CreateOrGet(ctx, c.Name, c.Phone)
			if err == nil && cust != nil {
				_, _ = mc.CustRepo.Update(ctx, cust.ID, bson.M{
					"tier":            c.Tier,
					"earnedPoints":    c.EarnedPoints,
					"redeemedPoints":  c.RedeemedPoints,
					"availablePoints": c.AvailablePoints,
					"totalSpent":      c.TotalSpent,
					"joinedDate":      c.JoinedDate,
				})
			}
		}
		if logger.Log != nil {
			logger.Log.Info().Int("count", len(customers)).Msg("👥 Seeded initial loyalty customers to MongoDB")
		}
	}
}
