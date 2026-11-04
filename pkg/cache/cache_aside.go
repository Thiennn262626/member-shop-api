package cache

import (
	"context"
	"time"
)

const (
	// Khóa Cache cho Menu và Sản phẩm công khai (Public APIs)
	KeyProductsActive = "shop:products:active" // Danh sách món đang mở bán (Storefront Menu)
	KeyProductsAll    = "shop:products:all"    // Toàn bộ sản phẩm cho Admin / POS
	PatternProducts   = "shop:product*"        // Pattern xóa khi có thay đổi sản phẩm

	// Khóa Cache cho Cấu hình Cửa hàng (Public API)
	KeyShopConfig = "shop:config:data" // Cấu hình tỷ lệ điểm, phí ship, giờ mở cửa

	// Khóa Cache cho Mã giảm giá (Vouchers)
	KeyVoucherPrefix = "shop:voucher:code:" // Tra cứu voucher theo code (shop:voucher:code:XYZ)
	KeyVouchersAll   = "shop:vouchers:all"  // Danh sách voucher
	PatternVouchers  = "shop:voucher*"      // Pattern xóa khi có thay đổi voucher

	// Khóa Cache cho Hội viên & Tích điểm
	KeyCustomerPhonePrefix = "shop:customer:phone:" // Tra cứu hội viên theo số điện thoại

	// Khóa Cache cho Ca trực thu ngân
	KeyCurrentShift = "shop:shift:current" // Ca trực hiện tại đang mở

	// Khóa Cache cho Xác thực & Token Blacklist (ECDSA JTI & Refresh Token Rotation)
	KeySessionRefreshTokenPrefix   = "auth:session:refresh:"   // auth:session:refresh:<token> -> session data
	KeyBlacklistRefreshTokenPrefix = "auth:blacklist:refresh:" // auth:blacklist:refresh:<token> -> rotated/revoked
	KeyBlacklistJTIPrefix          = "auth:blacklist:jti:"     // auth:blacklist:jti:<jti> -> revoked access token
	KeyAuthFailedAttemptsPrefix    = "auth:failed_pin:"        // auth:failed_pin:<username> -> đếm số lần nhập sai PIN để khóa tạm (SEC-05)
)

// CacheAside triển khai mô hình Cache-Aside tiêu chuẩn:
// 1. Kiểm tra cache trước. Nếu có dữ liệu hợp lệ (Cache Hit) -> trả về ngay, không truy vấn database.
// 2. Nếu không có (Cache Miss) -> gọi hàm queryDB() truy vấn dữ liệu gốc từ MongoDB.
// 3. Ghi kết quả mới vào cache với thời hạn TTL chỉ định, sau đó trả về dữ liệu.
// Hỗ trợ tự động fallback an toàn (nếu cache lỗi hoặc Redis offline, hệ thống vẫn hoạt động bình thường).
func CacheAside[T any](
	ctx context.Context,
	c RedisCacheService,
	key string,
	ttl time.Duration,
	queryDB func() (T, error),
) (T, error) {
	data, _, err := CacheAsideWithStatus(ctx, c, key, ttl, queryDB)
	return data, err
}

// CacheAsideWithStatus tương tự CacheAside nhưng trả về thêm cờ isHit (true nếu lấy từ Redis, false nếu miss từ DB).
func CacheAsideWithStatus[T any](
	ctx context.Context,
	c RedisCacheService,
	key string,
	ttl time.Duration,
	queryDB func() (T, error),
) (T, bool, error) {
	if c != nil {
		var cached T
		if err := c.Get(ctx, key, &cached); err == nil {
			return cached, true, nil
		}
	}

	data, err := queryDB()
	if err != nil {
		var zero T
		return zero, false, err
	}

	if c != nil {
		_ = c.Set(ctx, key, data, ttl)
	}

	return data, false, nil
}

// Invalidate xóa một hoặc nhiều khóa cache khi dữ liệu bị thay đổi (Mutation: Create / Update / Delete).
func Invalidate(ctx context.Context, c RedisCacheService, keys ...string) {
	if c == nil {
		return
	}
	for _, k := range keys {
		_ = c.Del(ctx, k)
	}
}

// InvalidatePattern xóa toàn bộ các khóa cache khớp với mẫu chỉ định (dùng SCAN/Pattern xóa an toàn).
func InvalidatePattern(ctx context.Context, c RedisCacheService, pattern string) {
	if c == nil {
		return
	}
	_ = c.Clear(ctx, pattern)
}
