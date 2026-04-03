package model

type RegisterRequest struct {
	Name     string `json:"name" binding:"required" example:"Budi Santoso"`
	Email    string `json:"email" binding:"required,email" example:"budi@company.com"`
	Password string `json:"password" binding:"required,min=6" example:"123456"`
	TenantID uint   `json:"tenant_id" binding:"required" example:"1"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"admin@yopmail.com"`
	Password string `json:"password" binding:"required" example:"123456"`
}

type LoginResponse struct {
	Token string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.xxx"`
	User  UserResponse `json:"user"`
}
