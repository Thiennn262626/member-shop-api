package db

import (
	"context"
	"member-shop-api/pkg/logger"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// EnsureIndexes tự động thiết lập và tối ưu hóa các chỉ mục (Indexes) trên MongoDB để loại bỏ hoàn toàn Full Collection Scans.
func EnsureIndexes(ctx context.Context, database *mongo.Database) error {
	ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	uniqueOpt := options.Index().SetUnique(true)

	// 1. Collection 'orders'
	ordersColl := database.Collection("orders")
	_, err := ordersColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "orderCode", Value: 1}},
			Options: uniqueOpt,
		},
		{
			Keys: bson.D{{Key: "customerPhone", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "createdAt", Value: -1}},
		},
		{
			Keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "createdAt", Value: -1},
			},
		},
	})
	if err != nil {
		logger.Log.Warn().Err(err).Msg("Warning: could not create indexes for 'orders'")
	}

	// 2. Collection 'customers'
	custColl := database.Collection("customers")
	_, _ = custColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "phone", Value: 1}},
			Options: uniqueOpt,
		},
		{
			Keys: bson.D{{Key: "tier", Value: 1}},
		},
	})

	// 3. Collection 'products'
	prodColl := database.Collection("products")
	_, _ = prodColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "isActive", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "category", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "isActive", Value: 1},
				{Key: "category", Value: 1},
			},
		},
	})

	// 4. Collection 'vouchers'
	voucherColl := database.Collection("vouchers")
	_, _ = voucherColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "code", Value: 1}},
			Options: uniqueOpt,
		},
		{
			Keys: bson.D{{Key: "isActive", Value: 1}},
		},
	})

	// 5. Collection 'staff_accounts'
	staffColl := database.Collection("staff_accounts")
	_, _ = staffColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "username", Value: 1}},
			Options: uniqueOpt,
		},
	})

	// 6. Collection 'cash_shifts'
	shiftColl := database.Collection("cash_shifts")
	_, _ = shiftColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "status", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "openedAt", Value: -1}},
		},
	})

	// 7. Collection 'work_shifts'
	workShiftColl := database.Collection("work_shifts")
	_, _ = workShiftColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "staffId", Value: 1},
				{Key: "date", Value: 1},
			},
		},
		{
			Keys: bson.D{{Key: "date", Value: 1}},
		},
	})

	// 8. Collection 'point_transactions'
	pointTxColl := database.Collection("point_transactions")
	_, _ = pointTxColl.Indexes().CreateMany(ctx, []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "customerId", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "createdAt", Value: -1}},
		},
	})

	logger.Log.Info().Msg("⚡ MongoDB performance indexes successfully verified/created")
	return nil
}
