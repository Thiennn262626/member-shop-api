package repository

import (
	"context"
	"errors"
	"member-shop-api/internal/model"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ConfigRepository interface {
	GetConfig(ctx context.Context) (*model.ShopConfig, error)
	UpdateConfig(ctx context.Context, updates bson.M) (*model.ShopConfig, error)
}

type configRepo struct {
	coll *mongo.Collection
}

func NewConfigRepository(db *mongo.Database) ConfigRepository {
	return &configRepo{coll: db.Collection("shop_configs")}
}

func (r *configRepo) GetConfig(ctx context.Context) (*model.ShopConfig, error) {
	var cfg model.ShopConfig
	err := r.coll.FindOne(ctx, bson.M{"_id": "current_config"}).Decode(&cfg)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Seed default config
			defaultCfg := model.ShopConfig{
				ID:                "current_config",
				ShopName:          "Bake & Sip",
				Tagline:           "Tiệm Bánh Nóng & Đồ Uống Tươi",
				Hotline:           "0901 234 567",
				ZaloPhone:         "0901234567",
				Address:           "68 Đường Hoa Sứ, Phường 7, Quận Phú Nhuận, TP.HCM",
				PointToVndRatio:   1000,
				SpendToPointRatio: 10000,
				IsOpen:            true,
				Bank: model.BankConfig{
					BankName:      "MBBank (Ngân Hàng Quân Đội)",
					AccountNumber: "0901234567",
					AccountHolder: "TIEM BANH BAKE AND SIP",
				},
				Delivery: model.DeliveryConfig{
					Enabled:       true,
					MaxDistanceKm: 10,
					AreaNote:      "Chỉ nhận giao hàng trong bán kính 10km quanh khu vực Phú Nhuận & lân cận",
					Tiers: []model.ShippingTier{
						{ID: "tier-1", MinOrder: 0, MaxOrder: int64Ptr(99999), Fee: 20000, Label: "Đơn dưới 100.000đ (Phí ship 20k)"},
						{ID: "tier-2", MinOrder: 100000, MaxOrder: int64Ptr(199999), Fee: 10000, Label: "Đơn từ 100.000đ - 199.999đ (Phí ship 10k)"},
						{ID: "tier-3", MinOrder: 200000, Fee: 0, Label: "Đơn từ 200.000đ trở lên (Freeship 0đ)"},
					},
				},
			}
			_, _ = r.coll.InsertOne(ctx, defaultCfg)
			return &defaultCfg, nil
		}
		return nil, err
	}
	return &cfg, nil
}

func (r *configRepo) UpdateConfig(ctx context.Context, updates bson.M) (*model.ShopConfig, error) {
	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(true)
	var cfg model.ShopConfig
	err := r.coll.FindOneAndUpdate(ctx, bson.M{"_id": "current_config"}, bson.M{"$set": updates}, opts).Decode(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

func int64Ptr(v int64) *int64 {
	return &v
}
