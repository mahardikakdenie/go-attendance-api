package model

import "time"

type RegisterRequest struct {
	Name        string `json:"name" binding:"required" example:"Budi Santoso"`
	Email       string `json:"email" binding:"required,email" example:"budi@company.com"`
	Password    string `json:"password" binding:"required,min=6" example:"123456"`
	TenantID    uint   `json:"tenant_id" binding:"required" example:"1"`
	RoleID      uint   `json:"role_id" example:"1"`
	Department  string `json:"department" example:"IT"`
	Address     string `json:"address" example:"Jl. Sudirman No. 1"`
	PhoneNumber string `json:"phone_number" example:"08123456789"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"admin@yopmail.com"`
	Password string `json:"password" binding:"required" example:"123456"`
}

type LoginResponse struct {
	Token string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.xxx"`
	User  UserResponse `json:"user"`
}

type Token struct {
	ID         uint `gorm:"primaryKey" json:"id"`
	UserID     uint `json:"user_id"`
	Token      string `gorm:"uniqueIndex" json:"-"`
	IPAddress  string `gorm:"type:varchar(50)" json:"ip_address"`
	UserAgent  string `gorm:"type:text" json:"user_agent"`
	DeviceInfo string `gorm:"type:varchar(255)" json:"device_info"`
	IsRevoked  bool   `gorm:"default:false" json:"is_revoked"`
	CreatedAt  time.Time `json:"created_at"`
	RevokedAt  *time.Time `json:"revoked_at"`
}

type SessionResponse struct {
	ID         uint      `json:"id"`
	IPAddress  string    `json:"ip_address"`
	UserAgent  string    `json:"user_agent"`
	DeviceInfo string    `json:"device_info"`
	IsActive   bool      `json:"is_active"`
	IsCurrent  bool      `json:"is_current"`
	LastActive time.Time `json:"last_active"`
}
