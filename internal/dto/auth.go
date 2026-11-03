package dto

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	PIN      string `json:"pin" binding:"required"`
}

type LoginResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	User         struct {
		ID       string `json:"id"`
		Username string `json:"username"`
		Name     string `json:"name"`
		Role     string `json:"role"`
	} `json:"user"`
}
