package model

import "time"

type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleManager  UserRole = "manager"
	RoleEmployee UserRole = "employee"
)

type User struct {
	ID       uint `gorm:"primaryKey" json:"id" example:"1"`
	TenantID uint `gorm:"index;not null" json:"tenant_id" example:"1"`

	Tenant *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`

	Name     string `gorm:"type:varchar(100);not null" json:"name" example:"Budi Santoso" binding:"required,min=3,max=100"`
	Email    string `gorm:"type:varchar(100);unique;not null" json:"email" example:"budi@company.com" binding:"required,email"`
	Password string `gorm:"type:varchar(255);not null" json:"-"` // tetap hidden

	Role UserRole `gorm:"type:varchar(50);default:employee" json:"role" example:"employee" binding:"omitempty,oneof=admin manager employee"`

	Attendances []Attendance `gorm:"foreignKey:UserID" json:"attendances,omitempty"`

	MediaUrl string `gorm:"type:varchar(255)" json:"media_url" example:"https://cdn.example.com/profile/budi.jpg" binding:"omitempty,url"`

	CreatedAt time.Time `json:"created_at" example:"2026-04-05T12:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2026-04-05T12:30:00Z"`
}

type UserFilter struct {
	Name     string
	Email    string
	Role     UserRole
	TenantID uint

	OrderBy string
	Sort    string

	Limit  int
	Offset int
}

type CreateUserRequest struct {
	Name     string   `json:"name" binding:"required" example:"Budi Santoso"`
	Email    string   `json:"email" binding:"required,email" example:"budi@company.com"`
	Password string   `json:"password" binding:"required,min=6" example:"123456"`
	Role     UserRole `json:"role" example:"employee"`
}

type UserResponse struct {
	ID        uint      `json:"id" example:"1"`
	Name      string    `json:"name" example:"Budi Santoso"`
	Email     string    `json:"email" example:"budi@company.com"`
	Role      UserRole  `json:"role" example:"employee"`
	TenantID  uint      `json:"tenant_id" example:"1"`
	CreatedAt time.Time `json:"created_at"`

	Tenant      *TenantResponse      `json:"tenant,omitempty"`
	Attendances []AttendanceResponse `json:"attendances,omitempty"`
}
