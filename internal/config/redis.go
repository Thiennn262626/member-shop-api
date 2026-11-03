package config

import (
	"context"
	"member-shop-api/internal/utils"
	"member-shop-api/pkg/logger"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Addr     string
	Username string
	Password string
	DB       int
}

func NewRedisClient() *redis.Client {
	addr := utils.GetEnv("REDIS_ADDR", utils.GetEnv("REDIS_HOST", "localhost:6379"))
	user := utils.GetEnv("REDIS_USER", utils.GetEnv("REDIS_USERNAME", ""))
	pass := utils.GetEnv("REDIS_PASSWORD", "")
	dbNum := utils.GetIntEnv("REDIS_DB", 0)

	cfg := RedisConfig{
		Addr:     addr,
		Username: user,
		Password: pass,
		DB:       dbNum,
	}

	client := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr,
		Username:     cfg.Username,
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     20,
		MinIdleConns: 5,
		DialTimeout:  3 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		logger.Log.Warn().Err(err).Str("addr", cfg.Addr).Msg("⚠️ Redis connection failed, running without Redis cache")
		return nil
	}

	logger.Log.Info().Str("addr", cfg.Addr).Msg("🍺 Connected Redis")
	return client
}
