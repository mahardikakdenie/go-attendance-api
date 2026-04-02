package model

import "time"

type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleManager  UserRole = "manager"
	RoleEmployee UserRole = "employee"
)

type User struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	TenantID uint `gorm:"index" json:"tenant_id"`

	Name     string `gorm:"type:varchar(100);not null" json:"name" example:"Budi Santoso"`
	Email    string `gorm:"type:varchar(100);unique;not null" json:"email" example:"budi@company.com"`
	Password string `gorm:"type:varchar(255);not null" json:"-" example:"123456"`

	Role UserRole `gorm:"type:varchar(50);default:employee" json:"role" example:"employee"`

	Attendances []Attendance `gorm:"foreignKey:UserID" json:"attendances,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserFilter struct {
	Name     string
	Email    string
	Role     UserRole
	TenantID uint

	OrderBy string
	Sort    string // asc / desc

	Limit  int
	Offset int
}

type CreateUserRequest struct {
	Name     string   `json:"name" binding:"required"`
	Email    string   `json:"email" binding:"required,email"`
	Password string   `json:"password" binding:"required,min=6"`
	Role     UserRole `json:"role"`
}

type UserResponse struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
}
