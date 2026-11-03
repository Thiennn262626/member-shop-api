package services

import (
	"context"
	"errors"
	"member-shop-api/internal/config"
	"member-shop-api/internal/dto"
	"member-shop-api/internal/model"
	"member-shop-api/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthService interface {
	Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error)
	GenerateTokens(user *model.StaffAccount) (string, string, error)
}

type authService struct {
	cfg       *config.Config
	staffRepo repository.StaffRepository
}

func NewAuthService(cfg *config.Config, staffRepo repository.StaffRepository) AuthService {
	return &authService{cfg: cfg, staffRepo: staffRepo}
}

func (s *authService) Login(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	staff, err := s.staffRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if staff == nil {
		return nil, errors.New("tên đăng nhập hoặc mã PIN không đúng")
	}

	if !staff.IsActive {
		return nil, errors.New("tài khoản nhân viên này đang bị tạm khóa")
	}

	if !staff.CheckPIN(req.PIN) {
		return nil, errors.New("tên đăng nhập hoặc mã PIN không đúng")
	}

	accessToken, refreshToken, err := s.GenerateTokens(staff)
	if err != nil {
		return nil, err
	}

	res := &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}
	res.User.ID = staff.ID.Hex()
	res.User.Username = staff.Username
	res.User.Name = staff.Name
	res.User.Role = string(staff.Role)

	return res, nil
}

func (s *authService) GenerateTokens(user *model.StaffAccount) (string, string, error) {
	now := time.Now()

	accessClaims := jwt.MapClaims{
		"sub":      user.ID.Hex(),
		"username": user.Username,
		"role":     user.Role,
		"name":     user.Name,
		"exp":      now.Add(time.Duration(s.cfg.AccessTokenExpires) * time.Second).Unix(),
		"iat":      now.Unix(),
	}
	accessToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims).SignedString([]byte(s.cfg.JWTAccessSecret))
	if err != nil {
		return "", "", err
	}

	refreshClaims := jwt.MapClaims{
		"sub": user.ID.Hex(),
		"exp": now.Add(time.Duration(s.cfg.RefreshTokenExpires) * time.Second).Unix(),
		"iat": now.Unix(),
	}
	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(s.cfg.JWTRefreshSecret))
	if err != nil {
		return "", "", err
	}

	return accessToken, refreshToken, nil
}
