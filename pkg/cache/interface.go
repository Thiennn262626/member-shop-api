package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisCacheService interface {
	GetClient() *redis.Client

	Get(ctx context.Context, key string, dest any) error
	GetTTL(ctx context.Context, key string) (time.Duration, error)
	Scan(ctx context.Context, key string) ([]string, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	Clear(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)

	Increment(ctx context.Context, key string, delta int64, ttl time.Duration) (int64, error)
	ZIncrement(ctx context.Context, key string, member string, score int64, ttl time.Duration) (float64, error)
	ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]ZMemberWithScore, error)
	ZAdd(ctx context.Context, key string, members ...redis.Z) error

	Rename(ctx context.Context, tmpKey string, key string) error

	HSet(ctx context.Context, hashKey string, key string, value any) error
	HGet(ctx context.Context, hashKey string, key string, dest any) error

	// Thay đổi: HMSet nhận vào map[string]any để lưu nhiều cặp field->value cùng lúc
	HMSet(ctx context.Context, hashKey string, values map[string]any) error
	// Thay đổi: dest của HMGet nhận vào một con trỏ slice []Struct hoặc []Type để hứng kết quả
	HMGet(ctx context.Context, hashKey string, keys []string, dest any) error
}
