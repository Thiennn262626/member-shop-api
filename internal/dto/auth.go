package dto

import "time"

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	PIN      string `json:"pin" binding:"required"`
}

type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	Token       string `json:"token"`
	ExpiresIn   int    `json:"expiresIn"` // Thời gian sống tính bằng giây (ví dụ: 900s = 15m)
	User        struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Name     string `json:"name"`
		Role     string `json:"role"`
	} `json:"user"`
}

type RefreshResponse struct {
	AccessToken string `json:"accessToken"`
	Token       string `json:"token"`
	ExpiresIn   int    `json:"expiresIn"`
}

// RefreshTokenSession lưu thông tin phiên làm việc của Refresh Token trong Redis/Cache
type RefreshTokenSession struct {
	UserID    string    `json:"userId"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}
