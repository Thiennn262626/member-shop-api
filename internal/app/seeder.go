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

	// 1. Seed Products if empty
	existingProducts, err := mc.ProductRepo.GetAll(ctx, false)
	if err == nil && len(existingProducts) == 0 {
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

		drinkSweetOptions := model.ItemOptionGroup{
			ID:        "drink-sweet",
			Name:      "Độ Ngọt",
			Required:  true,
			Selection: "only",
			Options: []model.ItemOptionItem{
				{ID: "sw-100", Name: "100% đường (Vừa miệng)", Price: 0, IsDefault: true, IsAvailable: true},
				{ID: "sw-50", Name: "50% đường (Ít ngọt thanh mát)", Price: 0, IsAvailable: true},
			},
		}

		drinkIceOptions := model.ItemOptionGroup{
			ID:        "drink-ice",
			Name:      "Lượng Đá",
			Required:  true,
			Selection: "only",
			Options: []model.ItemOptionItem{
				{ID: "ice-100", Name: "Đầy đá mát lạnh", Price: 0, IsDefault: true, IsAvailable: true},
				{ID: "ice-separate", Name: "Để đá riêng đi xa", Price: 0, IsAvailable: true},
			},
		}

		products := []model.Product{
			{
				Name:                 "Bánh Tráng Trộn Thập Cẩm Tây Ninh",
				Category:             "Bánh Tráng Trộn",
				Description:          "Bánh tráng dẻo mềm thấm muối tôm, khô bò đen, tép sấy, trứng cút lòng đào, xoài chua bào sợi, rau răm và hành phi thơm lừng.",
				Price:                20000,
				FakePrice:            25000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1546069901-ba9599a7e63c?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Bịch",
				Tag:                  "Bán Chạy 🔥",
				Rating:               4.9,
				SoldCount:            1250,
				OptionGroups:         []model.ItemOptionGroup{spiceOptions, customNoteOptions, toppingOptions},
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Bánh Tráng Cuốn Sốt Bơ Trứng Cút",
				Category:             "Bánh Tráng Cuốn",
				Description:          "Cuộn tròn ngập tràn khô bò, tép, hành phi, chấm đẫm sốt bơ béo ngậy vàng óng kèm sốt me chua ngọt đậm đà.",
				Price:                25000,
				FakePrice:            30000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1540420773420-3366772f4999?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Hộp 5 cuốn",
				Tag:                  "Signature ⭐",
				Rating:               5.0,
				SoldCount:            980,
				OptionGroups:         []model.ItemOptionGroup{spiceOptions, customNoteOptions, toppingOptions},
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Bánh Tráng Trộn Sốt Trứng Muối Khô Bò",
				Category:             "Bánh Tráng Trộn",
				Description:          "Bánh tráng dẻo thấm đẫm sốt trứng muối bùi béo mằn mặn, rắc đầy khô bò xé, tép khô và đậu phộng rang giòn bùi.",
				Price:                25000,
				FakePrice:            28000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1569058242253-92a9c755a0ec?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Bịch",
				Tag:                  "Đặc Biệt ✨",
				Rating:               4.9,
				SoldCount:            840,
				OptionGroups:         []model.ItemOptionGroup{spiceOptions, customNoteOptions, toppingOptions},
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Bánh Tráng Chấm Sốt Tắc Me Rim Cay",
				Category:             "Bánh Tráng Chấm",
				Description:          "Bánh tráng phơi sương dẻo quẹt sốt me tắc cay chua sánh đặc, rắc hành phi giòn tan ăn bao ghiền.",
				Price:                15000,
				FakePrice:            18000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1555126634-323283e090fa?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Phần",
				Tag:                  "Gây Nghiện 🌶️",
				Rating:               4.8,
				SoldCount:            720,
				OptionGroups:         []model.ItemOptionGroup{spiceOptions, toppingOptions},
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Bánh Tráng Nướng Mỡ Hành Trứng Cút",
				Category:             "Bánh Tráng Nướng",
				Description:          "Bánh tráng nướng than hoa nóng giòn rụm, quét mỡ hành thơm phức, trứng cút, xúc xích và tép giòn tan.",
				Price:                20000,
				FakePrice:            25000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1565299585323-38d6b0865b47?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Cái",
				Tag:                  "Nóng Giòn 🥐",
				Rating:               4.9,
				SoldCount:            610,
				OptionGroups:         []model.ItemOptionGroup{spiceOptions},
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Bánh Tráng Xì Ke Muối Tỏi Tây Ninh",
				Category:             "Bánh Tráng Trộn",
				Description:          "Bánh tráng siêu mỏng thấm muối tỏi ớt cay tê tái, càng nhai càng dẻo thơm ngất ngây, ăn là không dứt được.",
				Price:                15000,
				FakePrice:            18000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1563379091339-03b21ab4a4f8?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Bịch",
				Tag:                  "Hot Trend ⚡",
				Rating:               4.8,
				SoldCount:            1100,
				OptionGroups:         []model.ItemOptionGroup{spiceOptions},
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Bánh Tráng Cuốn Sốt Me Cay Chua Ngọt",
				Category:             "Bánh Tráng Cuốn",
				Description:          "Cuốn dẻo mềm nhân xoài tép hành phi, rưới sốt me đặc sánh cay chua đậm đà kích thích vị giác.",
				Price:                20000,
				FakePrice:            23000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1540420773420-3366772f4999?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Hộp 4 cuốn",
				Tag:                  "Bán Chạy 🔥",
				Rating:               4.9,
				SoldCount:            550,
				OptionGroups:         []model.ItemOptionGroup{spiceOptions, customNoteOptions, toppingOptions},
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Cóc Non Lắc Muối Tôm Bò Khô",
				Category:             "Đồ Ăn Vặt",
				Description:          "Cóc non giòn sần sật lắc đều muối tôm cay nồng, đường thốt nốt và khô bò vụn.",
				Price:                20000,
				FakePrice:            25000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1568901346375-23c9450c58cd?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Hộp",
				Tag:                  "Chua Cay 🤤",
				Rating:               4.8,
				SoldCount:            430,
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Xoài Non Lắc Mắm Ruốc Cay",
				Category:             "Đồ Ăn Vặt",
				Description:          "Xoài keo giòn sần sật ngâm sốt mắm ruốc xào cay thơm đậm vị miền Tây ăn là dính.",
				Price:                20000,
				FakePrice:            25000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1550547660-d9450f859349?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Hộp",
				Tag:                  "Mới Ra Mắt ✨",
				Rating:               4.9,
				SoldCount:            390,
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Trà Tắc Khổng Lồ Mát Lạnh Giải Cay",
				Category:             "Đồ Uống Giải Khát",
				Description:          "Ly khổng lồ 700ml trà lài ủ lạnh vắt tắc tươi chua thanh mát giải cay tức thì khi ăn bánh tráng.",
				Price:                15000,
				FakePrice:            18000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1556679343-c7306c1976bc?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Ly 700ml",
				Tag:                  "Combo Chuẩn 🧋",
				Rating:               4.9,
				SoldCount:            1400,
				OptionGroups:         []model.ItemOptionGroup{drinkSweetOptions, drinkIceOptions},
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Trà Dâu Tằm Tươi Lắc Đá",
				Category:             "Đồ Uống Giải Khát",
				Description:          "Vị dâu tằm chua ngọt hòa quyện trà thơm ngát, topping lát dâu giòn rụm sảng khoái.",
				Price:                20000,
				FakePrice:            25000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1558857563-b37cf5a133df?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Ly",
				Tag:                  "Signature ⭐",
				Rating:               4.9,
				SoldCount:            890,
				OptionGroups:         []model.ItemOptionGroup{drinkSweetOptions, drinkIceOptions},
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Trà Đào Dầm Miếng Giòn",
				Category:             "Đồ Uống Giải Khát",
				Description:          "Trà đào thanh nhẹ kết hợp cam tươi và 2 miếng đào giòn ngọt thơm lừng.",
				Price:                20000,
				FakePrice:            25000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1513558161293-cdaf765ed2fd?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Ly",
				Tag:                  "Bán Chạy 🔥",
				Rating:               4.8,
				SoldCount:            670,
				OptionGroups:         []model.ItemOptionGroup{drinkSweetOptions, drinkIceOptions},
				CreatedAt:            now,
				UpdatedAt:            now,
			},
		}

		for _, p := range products {
			_ = mc.ProductRepo.Create(ctx, &p)
		}
		if logger.Log != nil {
			logger.Log.Info().Int("count", len(products)).Msg("🌶️ Seeded initial Bánh Tráng products to MongoDB")
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
