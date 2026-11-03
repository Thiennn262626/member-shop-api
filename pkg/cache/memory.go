package cache

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrCacheMiss = redis.Nil

type memItem struct {
	data      []byte
	expiresAt time.Time
}

func (i memItem) isExpired() bool {
	if i.expiresAt.IsZero() {
		return false
	}
	return time.Now().After(i.expiresAt)
}

type memoryCacheService struct {
	mu    sync.RWMutex
	items map[string]memItem
}

// NewMemoryCacheService tạo cache in-memory dự phòng tự động khi không có kết nối Redis.
func NewMemoryCacheService() RedisCacheService {
	s := &memoryCacheService{
		items: make(map[string]memItem),
	}

	// Chạy định kỳ dọn dẹp các key hết hạn trong RAM
	go func() {
		ticker := time.NewTicker(2 * time.Minute)
		for range ticker.C {
			s.mu.Lock()
			now := time.Now()
			for k, v := range s.items {
				if !v.expiresAt.IsZero() && now.After(v.expiresAt) {
					delete(s.items, k)
				}
			}
			s.mu.Unlock()
		}
	}()

	return s
}

func (m *memoryCacheService) GetClient() *redis.Client {
	return nil
}

func (m *memoryCacheService) Get(ctx context.Context, key string, dest any) error {
	m.mu.RLock()
	item, ok := m.items[key]
	m.mu.RUnlock()

	if !ok || item.isExpired() {
		return ErrCacheMiss
	}

	return json.Unmarshal(item.data, dest)
}

func (m *memoryCacheService) GetTTL(ctx context.Context, key string) (time.Duration, error) {
	m.mu.RLock()
	item, ok := m.items[key]
	m.mu.RUnlock()

	if !ok || item.isExpired() {
		return 0, ErrCacheMiss
	}

	if item.expiresAt.IsZero() {
		return -1, nil
	}

	return time.Until(item.expiresAt), nil
}

func (m *memoryCacheService) Scan(ctx context.Context, pattern string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var matched []string
	for k, item := range m.items {
		if item.isExpired() {
			continue
		}
		if matchPattern(pattern, k) {
			matched = append(matched, k)
		}
	}
	return matched, nil
}

func (m *memoryCacheService) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	m.mu.Lock()
	m.items[key] = memItem{
		data:      data,
		expiresAt: expiresAt,
	}
	m.mu.Unlock()

	return nil
}

func (m *memoryCacheService) Del(ctx context.Context, key string) error {
	m.mu.Lock()
	delete(m.items, key)
	m.mu.Unlock()
	return nil
}

func (m *memoryCacheService) Clear(ctx context.Context, pattern string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for k := range m.items {
		if matchPattern(pattern, k) {
			delete(m.items, k)
		}
	}
	return nil
}

func (m *memoryCacheService) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.RLock()
	item, ok := m.items[key]
	m.mu.RUnlock()

	if !ok || item.isExpired() {
		return false, nil
	}
	return true, nil
}

func (m *memoryCacheService) Increment(ctx context.Context, key string, delta int64, ttl time.Duration) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var val int64
	if item, ok := m.items[key]; ok && !item.isExpired() {
		_ = json.Unmarshal(item.data, &val)
	}

	val += delta
	data, _ := json.Marshal(val)

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	m.items[key] = memItem{
		data:      data,
		expiresAt: expiresAt,
	}
	return val, nil
}

func (m *memoryCacheService) ZIncrement(ctx context.Context, key string, member string, score int64, ttl time.Duration) (float64, error) {
	return float64(score), nil
}

func (m *memoryCacheService) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]ZMemberWithScore, error) {
	return nil, nil
}

func (m *memoryCacheService) ZAdd(ctx context.Context, key string, members ...redis.Z) error {
	return nil
}

func (m *memoryCacheService) Rename(ctx context.Context, tmpKey string, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if item, ok := m.items[tmpKey]; ok {
		m.items[key] = item
		delete(m.items, tmpKey)
		return nil
	}
	return errors.New("key not found")
}

func (m *memoryCacheService) HSet(ctx context.Context, hashKey string, key string, value any) error {
	return m.Set(ctx, hashKey+":"+key, value, 0)
}

func (m *memoryCacheService) HGet(ctx context.Context, hashKey string, key string, dest any) error {
	return m.Get(ctx, hashKey+":"+key, dest)
}

func (m *memoryCacheService) HMSet(ctx context.Context, hashKey string, values map[string]any) error {
	for k, v := range values {
		if err := m.HSet(ctx, hashKey, k, v); err != nil {
			return err
		}
	}
	return nil
}

func (m *memoryCacheService) HMGet(ctx context.Context, hashKey string, keys []string, dest any) error {
	return nil
}

func matchPattern(pattern, key string) bool {
	if pattern == "*" || pattern == "" {
		return true
	}
	// Hỗ trợ match tiền tố đơn giản dạng "products:*" hoặc wildcard filepath
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		if strings.HasPrefix(key, prefix) {
			return true
		}
	}
	matched, _ := filepath.Match(pattern, key)
	return matched
}
