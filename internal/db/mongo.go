package db

import (
	"context"
	"fmt"
	"member-shop-api/internal/config"
	"member-shop-api/pkg/logger"
	"time"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDB struct {
	Client   *mongo.Client
	Database *mongo.Database
}

func ConnectMongoDB(cfg *config.Config) (*MongoDB, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Tối ưu hóa Connection Pool cho MongoDB để xử lý tải cao, duy trì kết nối sẵn sàng (keep-alive)
	clientOpts := options.Client().
		ApplyURI(cfg.MongoURI).
		SetMaxPoolSize(100).
		SetMinPoolSize(10).
		SetMaxConnIdleTime(5 * time.Minute).
		SetConnectTimeout(10 * time.Second)

	client, err := mongo.Connect(clientOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to create mongo client: %w", err)
	}

	// Ping database
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongodb: %w", err)
	}

	database := client.Database(cfg.MongoDB)
	logger.Log.Info().Str("database", cfg.MongoDB).Msg("✅ Connected to MongoDB successfully")

	// Đảm bảo và tạo các chỉ mục hiệu năng (Indexes)
	_ = EnsureIndexes(context.Background(), database)

	return &MongoDB{
		Client:   client,
		Database: database,
	}, nil
}

func (m *MongoDB) Close(ctx context.Context) error {
	if m.Client != nil {
		return m.Client.Disconnect(ctx)
	}
	return nil
}
