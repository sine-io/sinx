package dto

type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50" example:"john_doe"`
	Email    string `json:"email" binding:"required,email,max=100" example:"john@example.com"`
	Password string `json:"password" binding:"required,min=6,max=50" example:"password123"`
}

type LoginRequest struct {
	Username string `json:"username" binding:"required" example:"john_doe"`
	Password string `json:"password" binding:"required" example:"password123"`
}

type UserResponse struct {
	ID       uint   `json:"id" example:"1"`
	Username string `json:"username" example:"john_doe"`
	Email    string `json:"email" example:"john@example.com"`
	IsActive bool   `json:"is_active" example:"true"`
}

type LoginResponse struct {
	Token string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User  UserResponse `json:"user"`
}
