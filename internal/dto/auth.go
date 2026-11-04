package dto

import "time"

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	PIN      string `json:"pin" binding:"required"`
}

type AdminUserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Name      string `json:"name"`
	Role      string `json:"role"`
	RoleTitle string `json:"roleTitle"`
	Shift     string `json:"shift,omitempty"`
}

type LoginResponse struct {
	AccessToken string            `json:"accessToken"`
	ExpiresIn   int               `json:"expiresIn"` // Thời gian sống tính bằng giây (ví dụ: 900s = 15m)
	User        AdminUserResponse `json:"user"`
}

type RefreshResponse struct {
	AccessToken string `json:"accessToken"`
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
