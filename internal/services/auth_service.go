package services

import (
	"context"
	"member-shop-api/internal/config"
	"member-shop-api/internal/dto"
	"member-shop-api/internal/repository"
	"member-shop-api/pkg/apperror"
	"member-shop-api/pkg/authjwt"
	"member-shop-api/pkg/cache"
	"member-shop-api/pkg/logger"
	"net/http"
	"time"
)

type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, string, error)
	RefreshToken(ctx context.Context, oldRefreshToken string) (*dto.RefreshResponse, string, error)
	Logout(ctx context.Context, accessToken string, refreshToken string) error
	GetKeyManager() *authjwt.ECDSAKeyManager
}

type authService struct {
	cfg        *config.Config
	staffRepo  repository.StaffRepository
	cache      cache.RedisCacheService
	keyManager *authjwt.ECDSAKeyManager
}

func NewAuthService(cfg *config.Config, staffRepo repository.StaffRepository, cacheSvc cache.RedisCacheService) AuthService {
	km, err := authjwt.NewECDSAKeyManager(cfg.JWTECDSAPrivateKey, cfg.JWTECDSAPublicKey)
	if err != nil {
		if logger.Log != nil {
			logger.Log.Fatal().Err(err).Msg("Failed to initialize ECDSA key manager")
		}
	}

	return &authService{
		cfg:        cfg,
		staffRepo:  staffRepo,
		cache:      cacheSvc,
		keyManager: km,
	}
}

func (s *authService) GetKeyManager() *authjwt.ECDSAKeyManager {
	return s.keyManager
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, string, error) {
	staff, err := s.staffRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, "", apperror.Internal(err, "Không thể truy vấn thông tin nhân viên")
	}
	if staff == nil {
		return nil, "", apperror.New(http.StatusUnauthorized, apperror.CodeAuthInvalidCredentials, "Tên đăng nhập hoặc mã PIN không chính xác")
	}

	if !staff.IsActive {
		return nil, "", apperror.New(http.StatusForbidden, apperror.CodeAuthAccountLocked, "Tài khoản nhân viên này đang bị tạm khóa")
	}

	if !staff.CheckPIN(req.PIN) {
		return nil, "", apperror.New(http.StatusUnauthorized, apperror.CodeAuthInvalidCredentials, "Tên đăng nhập hoặc mã PIN không chính xác")
	}

	// 1. Tạo Access Token bằng thuật toán ECDSA ES256 với thời gian ngắn (mặc định 15 phút)
	accessTTL := time.Duration(s.cfg.AccessTokenExpires) * time.Second
	accessToken, _, _, err := s.keyManager.SignAccessToken(
		staff.ID.Hex(),
		staff.Username,
		string(staff.Role),
		staff.Name,
		accessTTL,
	)
	if err != nil {
		return nil, "", apperror.Internal(err, "Không thể ký ECDSA Access Token")
	}

	// 2. Tạo Refresh Token là một chuỗi ngẫu nhiên bảo mật (Opaque string 64 hex chars)
	refreshToken, err := authjwt.GenerateRandomRefreshToken()
	if err != nil {
		return nil, "", apperror.Internal(err, "Không thể tạo Refresh Token")
	}

	// 3. Lưu thông tin phiên làm việc của Refresh Token vào Redis/Cache
	refreshTTL := time.Duration(s.cfg.RefreshTokenExpires) * time.Second
	session := dto.RefreshTokenSession{
		UserID:    staff.ID.Hex(),
		Username:  staff.Username,
		Role:      string(staff.Role),
		Name:      staff.Name,
		CreatedAt: time.Now().UTC(),
	}
	sessionKey := cache.KeySessionRefreshTokenPrefix + refreshToken
	if err := s.cache.Set(ctx, sessionKey, session, refreshTTL); err != nil {
		return nil, "", apperror.Internal(err, "Không thể lưu phiên đăng nhập")
	}

	res := &dto.LoginResponse{
		AccessToken: accessToken,
		Token:       accessToken,
		ExpiresIn:   s.cfg.AccessTokenExpires,
	}
	res.User.ID = staff.ID.Hex()
	res.User.Username = staff.Username
	res.User.Name = staff.Name
	res.User.Role = string(staff.Role)

	// Trả về response JSON và chuỗi refreshToken riêng để Handler set vào HTTP-Only Cookie
	return res, refreshToken, nil
}

func (s *authService) RefreshToken(ctx context.Context, oldRefreshToken string) (*dto.RefreshResponse, string, error) {
	if oldRefreshToken == "" {
		return nil, "", apperror.New(http.StatusUnauthorized, apperror.CodeUnauthorized, "Yêu cầu cung cấp Refresh Token trong cookie")
	}

	// 1. Kiểm tra xem Refresh Token cũ có nằm trong Blacklist hay không (Phát hiện Replay Attack)
	blacklistKey := cache.KeyBlacklistRefreshTokenPrefix + oldRefreshToken
	isBlacklisted, _ := s.cache.Exists(ctx, blacklistKey)
	if isBlacklisted {
		if logger.Log != nil {
			logger.Log.Warn().Str("token_prefix", oldRefreshToken[:8]).Msg("⚠️ Detected attempt to reuse revoked/blacklisted Refresh Token (Replay Attack)")
		}
		return nil, "", apperror.New(http.StatusUnauthorized, apperror.CodeAuthTokenInvalid, "Refresh token đã bị thu hồi hoặc đã được sử dụng")
	}

	// 2. Lấy thông tin phiên làm việc từ Redis/Cache
	sessionKey := cache.KeySessionRefreshTokenPrefix + oldRefreshToken
	var session dto.RefreshTokenSession
	err := s.cache.Get(ctx, sessionKey, &session)
	if err != nil {
		return nil, "", apperror.New(http.StatusUnauthorized, apperror.CodeAuthTokenExpired, "Phiên đăng nhập đã hết hạn hoặc không tồn tại, vui lòng đăng nhập lại")
	}

	// 3. THU HỒI TOKEN CŨ: Xóa session cũ và ĐƯA REFRESH TOKEN CŨ VÀO BLACKLIST
	_ = s.cache.Del(ctx, sessionKey)
	// Lưu vào Blacklist với thời hạn 24 giờ để chặn mọi hành vi tái sử dụng
	_ = s.cache.Set(ctx, blacklistKey, "rotated", 24*time.Hour)

	// 4. Tạo cặp Access Token ECDSA mới và Refresh Token ngẫu nhiên mới (Token Rotation)
	accessTTL := time.Duration(s.cfg.AccessTokenExpires) * time.Second
	newAccessToken, _, _, err := s.keyManager.SignAccessToken(
		session.UserID,
		session.Username,
		session.Role,
		session.Name,
		accessTTL,
	)
	if err != nil {
		return nil, "", apperror.Internal(err, "Không thể ký ECDSA Access Token mới")
	}

	newRefreshToken, err := authjwt.GenerateRandomRefreshToken()
	if err != nil {
		return nil, "", apperror.Internal(err, "Không thể tạo Refresh Token mới")
	}

	// 5. Lưu phiên làm việc cho Refresh Token mới
	newSessionKey := cache.KeySessionRefreshTokenPrefix + newRefreshToken
	refreshTTL := time.Duration(s.cfg.RefreshTokenExpires) * time.Second
	session.CreatedAt = time.Now().UTC()
	if err := s.cache.Set(ctx, newSessionKey, session, refreshTTL); err != nil {
		return nil, "", apperror.Internal(err, "Không thể lưu phiên đăng nhập mới")
	}

	res := &dto.RefreshResponse{
		AccessToken: newAccessToken,
		Token:       newAccessToken,
		ExpiresIn:   s.cfg.AccessTokenExpires,
	}

	return res, newRefreshToken, nil
}

func (s *authService) Logout(ctx context.Context, accessToken string, refreshToken string) error {
	// 1. Đưa Access Token JTI vào Blacklist cho đến khi hết hạn tự nhiên
	if accessToken != "" {
		claims, err := s.keyManager.ValidateAccessToken(accessToken)
		if err == nil && claims != nil && claims.JTI != "" {
			var remaining time.Duration = 15 * time.Minute
			if claims.ExpiresAt != nil {
				rem := time.Until(claims.ExpiresAt.Time)
				if rem > 0 {
					remaining = rem
				}
			}
			jtiKey := cache.KeyBlacklistJTIPrefix + claims.JTI
			_ = s.cache.Set(ctx, jtiKey, "revoked", remaining)

			if logger.Log != nil {
				logger.Log.Info().Str("jti", claims.JTI).Dur("ttl", remaining).Msg("🚫 Access token JTI added to blacklist")
			}
		}
	}

	// 2. Thu hồi Refresh Token: Xóa session và đưa vào Blacklist
	if refreshToken != "" {
		sessionKey := cache.KeySessionRefreshTokenPrefix + refreshToken
		_ = s.cache.Del(ctx, sessionKey)

		blacklistKey := cache.KeyBlacklistRefreshTokenPrefix + refreshToken
		_ = s.cache.Set(ctx, blacklistKey, "logged_out", 24*time.Hour)

		if logger.Log != nil {
			logger.Log.Info().Msg("🚫 Refresh token invalidated and added to blacklist")
		}
	}

	return nil
}
