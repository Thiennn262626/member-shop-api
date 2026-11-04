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
		products := []model.Product{
			{
				Name:                 "Bánh Croissant Bơ Pháp Nướng Giòn",
				Category:             "Bánh Nướng Nóng",
				Description:          "Bột ngàn lớp chuẩn Pháp thơm lừng bơ Échiré, nướng vỏ giòn rụm, ruột mềm mướt thơm ngát.",
				Price:                35000,
				FakePrice:            42000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1555507036-ab1f4038808a?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Cái",
				Tag:                  "Signature ⭐",
				Rating:               4.9,
				SoldCount:            380,
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Bánh Mì Bơ Tỏi Phô Mai Chảy",
				Category:             "Bánh Mì & Mặn",
				Description:          "Vỏ bánh mì mềm xốp nhúng ngập sốt bơ tỏi mật ong thơm béo và kem phô mai chảy ngập tràn.",
				Price:                38000,
				FakePrice:            45000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1509722747041-616f39b57569?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Ổ",
				Tag:                  "Bán Chạy 🔥",
				Rating:               4.8,
				SoldCount:            520,
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Bánh Tart Trứng Nướng Than Caramel",
				Category:             "Bánh Nướng Nóng",
				Description:          "Đế tart ngàn lớp giòn tan bọc lớp kem trứng béo mịn nướng xém cạnh caramel ngọt thanh quyến rũ.",
				Price:                25000,
				FakePrice:            28000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1587314168485-3236d6710814?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Cái",
				Tag:                  "Mới Ra Lò 🥐",
				Rating:               4.9,
				SoldCount:            290,
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Bánh Tiramisu Kem Phô Mai Ý",
				Category:             "Bánh Ngọt & Kem",
				Description:          "Bánh ladyfinger đẫm cà phê espresso thơm nồng, quyện lớp kem mascarpone béo ngậy phủ bột cacao xịn.",
				Price:                49000,
				FakePrice:            55000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1571877227200-a0d98ea607e9?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Hộp",
				Tag:                  "Bán Chạy 🔥",
				Rating:               5.0,
				SoldCount:            410,
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Bánh Su Kem Vani Hokkaido",
				Category:             "Bánh Ngọt & Kem",
				Description:          "Vỏ choux giòn rụm bên ngoài, nhân kem sữa tươi Hokkaido mát lạnh, ngọt thanh, cắn ngập miệng.",
				Price:                29000,
				FakePrice:            32000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1509440159596-0249088772ff?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Hộp 2 cái",
				Tag:                  "Mới Ra Lò 🥐",
				Rating:               4.7,
				SoldCount:            195,
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Bánh Mì Sandwich Gà Xé Phô Mai Nướng",
				Category:             "Bánh Mì & Mặn",
				Description:          "Bánh sandwich nướng giòn kẹp thịt gà xé đậm vị sốt Mayo cay nhẹ và phô mai mozzarella kéo sợi.",
				Price:                36000,
				FakePrice:            40000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1528735602780-2552fd46c7af?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Phần",
				Tag:                  "Giảm Giá ⚡",
				Rating:               4.8,
				SoldCount:            230,
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Trà Sữa Ô Long Nướng Trân Châu Dẻo",
				Category:             "Trà & Trà Sữa",
				Description:          "Vị trà ô long sao cháy thơm sâu đậm hòa cùng sữa béo dịu và trân châu hoàng kim dai ngon.",
				Price:                39000,
				FakePrice:            45000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1558857563-b37cf5a133df?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Ly",
				Tag:                  "Bán Chạy 🔥",
				Rating:               4.9,
				SoldCount:            650,
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Trà Đào Cam Sả Tươi Mát Lạnh",
				Category:             "Trà & Trà Sữa",
				Description:          "Trà đào thanh nhẹ kết hợp cam tươi mọng nước, sả thơm thư giãn và 3 miếng đào giòn ngọt.",
				Price:                35000,
				FakePrice:            40000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1556679343-c7306c1976bc?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Ly",
				Tag:                  "Signature ⭐",
				Rating:               4.8,
				SoldCount:            480,
				CreatedAt:            now,
				UpdatedAt:            now,
			},
			{
				Name:                 "Cà Phê Muối Kem Béo Đậm Vị",
				Category:             "Cà Phê & Đá Xay",
				Description:          "Cà phê Robusta thơm đậm nguyên chất phủ lớp kem béo mặn mòi, khuấy đều uống cực cuốn.",
				Price:                32000,
				FakePrice:            38000,
				AllowPointRedemption: true,
				ImageURL:             "https://images.unsplash.com/photo-1514432324607-a09d9b4aefdd?w=500&auto=format&fit=crop&q=80",
				IsActive:             true,
				Unit:                 "Ly",
				Tag:                  "Signature ⭐",
				Rating:               4.9,
				SoldCount:            560,
				CreatedAt:            now,
				UpdatedAt:            now,
			},
		}

		for _, p := range products {
			_ = mc.ProductRepo.Create(ctx, &p)
		}
		if logger.Log != nil {
			logger.Log.Info().Int("count", len(products)).Msg("🥐 Seeded initial menu products to MongoDB")
		}
	}

	// 2. Seed Vouchers if empty
	existingVouchers, err := mc.VoucherRepo.GetAll(ctx)
	if err == nil && len(existingVouchers) == 0 {
		startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
		endDate := time.Date(2026, 12, 31, 23, 59, 59, 0, time.UTC)
		vouchers := []model.Voucher{
			{
				Code:              "BAKESIP10",
				Title:             "Giảm 10% Tối Đa 30k (Đơn từ 80k)",
				DiscountType:      model.VoucherDiscountPercent,
				DiscountValue:     10,
				MaxDiscountAmount: 30000,
				MinOrderValue:     80000,
				StartDate:         startDate,
				EndDate:           endDate,
				UsageLimit:        500,
				UsedCount:         0,
				IsActive:          true,
				CreatedAt:         time.Now(),
			},
			{
				Code:          "BANHNGON20K",
				Title:         "Giảm Ngay 20.000đ (Đơn từ 120k)",
				DiscountType:  model.VoucherDiscountFixed,
				DiscountValue: 20000,
				MinOrderValue: 120000,
				StartDate:     startDate,
				EndDate:       endDate,
				UsageLimit:    300,
				UsedCount:     0,
				IsActive:      true,
				CreatedAt:     time.Now(),
			},
			{
				Code:          "FREESHIP",
				Title:         "Hỗ Trợ Phí Ship 20.000đ (Đơn từ 100k)",
				DiscountType:  model.VoucherDiscountFixed,
				DiscountValue: 20000,
				MinOrderValue: 100000,
				StartDate:     startDate,
				EndDate:       endDate,
				UsageLimit:    200,
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
				EarnedPoints:    480,
				RedeemedPoints:  60,
				AvailablePoints: 420,
				TotalSpent:      3450000,
				JoinedDate:      "15/12/2025",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			{
				Phone:           "0977333444",
				Name:            "Minh Quân (Quân Bùi)",
				Tier:            model.TierGold,
				EarnedPoints:    175,
				RedeemedPoints:  25,
				AvailablePoints: 150,
				TotalSpent:      1450000,
				JoinedDate:      "10/01/2026",
				CreatedAt:       time.Now(),
				UpdatedAt:       time.Now(),
			},
			{
				Phone:           "0912345678",
				Name:            "Phương Thảo",
				Tier:            model.TierSilver,
				EarnedPoints:    60,
				RedeemedPoints:  15,
				AvailablePoints: 45,
				TotalSpent:      420000,
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
