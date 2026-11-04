package repository

import (
	"context"
	"errors"
	"fmt"
	"member-shop-api/internal/model"
	"member-shop-api/pkg/apperror"
	"net/http"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type ConfigRepository interface {
	GetConfig(ctx context.Context) (*model.ShopConfig, error)
	UpdateConfig(ctx context.Context, updates bson.M, expectedVersion ...int64) (*model.ShopConfig, error)
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
				ShopName:          "Bánh Tráng Cậu Hai",
				Tagline:           "Bánh Tráng Phơi Sương Chuẩn Vị Tây Ninh • Cay Béo Xuýt Xoa • Càng Ăn Càng Cuốn",
				Hotline:           "0334395313",
				ZaloPhone:         "0334395313",
				Address:           "7C Đ. Tô Hiến Thành, Nha Trang, Khánh Hòa",
				PointToVndRatio:   500,
				SpendToPointRatio: 10000,
				IsOpen:            true,
				Bank: model.BankConfig{
					BankName:      "MBBank (Ngân Hàng Quân Đội)",
					AccountNumber: "0334395313",
					AccountHolder: "NGUYEN DINH THIEN",
				},
				Delivery: model.DeliveryConfig{
					Enabled:       true,
					MaxDistanceKm: 10,
					AreaNote:      "Giao nhanh quanh khu vực Nha Trang",
					Tiers: []model.ShippingTier{
						{ID: "tier-1", MinOrder: 0, MaxOrder: int64Ptr(49999), Fee: 15000, Label: "Đơn dưới 50.000đ (Phí ship 15k)"},
						{ID: "tier-2", MinOrder: 50000, MaxOrder: int64Ptr(99999), Fee: 10000, Label: "Đơn từ 50.000đ - 99.999đ (Phí ship 10k)"},
						{ID: "tier-3", MinOrder: 150000, Fee: 0, Label: "Đơn từ 150.000đ trở lên (Freeship 0đ)"},
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

func (r *configRepo) UpdateConfig(ctx context.Context, updates bson.M, expectedVersion ...int64) (*model.ShopConfig, error) {
	delete(updates, "_id")
	delete(updates, "id")
	delete(updates, "version")

	filter := bson.M{"_id": "current_config"}
	hasVersionCheck := len(expectedVersion) > 0 && expectedVersion[0] >= 0
	if hasVersionCheck {
		expVer := expectedVersion[0]
		filter["$or"] = []bson.M{
			{"version": expVer},
			{"version": bson.M{"$exists": false}},
		}
	}

	updateDoc := bson.M{
		"$set": updates,
		"$inc": bson.M{"version": 1},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After).SetUpsert(false)
	var cfg model.ShopConfig
	err := r.coll.FindOneAndUpdate(ctx, filter, updateDoc, opts).Decode(&cfg)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			existing, checkErr := r.GetConfig(ctx)
			if checkErr == nil && existing != nil && hasVersionCheck {
				return nil, apperror.VersionMismatch(fmt.Sprintf("Cấu hình tiệm đã được cập nhật bởi một thiết bị khác (phiên bản mới: %d). Hệ thống đang tải lại dữ liệu mới nhất.", existing.Version))
			}
			return nil, apperror.New(http.StatusNotFound, apperror.CodeNotFound, "Không tìm thấy cấu hình tiệm")
		}
		return nil, err
	}
	return &cfg, nil
}

func int64Ptr(v int64) *int64 {
	return &v
}
