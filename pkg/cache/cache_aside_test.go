package cache

import (
	"context"
	"testing"
	"time"
)

type TestItem struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestCacheAsideFlow(t *testing.T) {
	ctx := context.Background()
	c := NewMemoryCacheService()

	dbCallCount := 0
	fetchFromDB := func() (*TestItem, error) {
		dbCallCount++
		return &TestItem{ID: "1", Name: "Trà Đào"}, nil
	}

	key := "test:item:1"
	ttl := 1 * time.Minute

	// 1st Call: Cache MISS -> Must invoke fetchFromDB
	item, err := CacheAside(ctx, c, key, ttl, fetchFromDB)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if item.Name != "Trà Đào" {
		t.Errorf("expected item name 'Trà Đào', got %s", item.Name)
	}
	if dbCallCount != 1 {
		t.Errorf("expected dbCallCount 1 on cache miss, got %d", dbCallCount)
	}

	// 2nd Call: Cache HIT -> Must NOT invoke fetchFromDB
	item2, err := CacheAside(ctx, c, key, ttl, fetchFromDB)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if item2.Name != "Trà Đào" {
		t.Errorf("expected item name 'Trà Đào', got %s", item2.Name)
	}
	if dbCallCount != 1 {
		t.Errorf("expected dbCallCount to remain 1 on cache hit, got %d", dbCallCount)
	}

	// 3rd Call: Invalidate key -> Next call must be Cache MISS
	Invalidate(ctx, c, key)
	_, _ = CacheAside(ctx, c, key, ttl, fetchFromDB)
	if dbCallCount != 2 {
		t.Errorf("expected dbCallCount 2 after invalidation, got %d", dbCallCount)
	}
}
