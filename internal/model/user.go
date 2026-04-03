package model

import "time"

// ======================
// USER ROLE (ENUM)
// ======================
type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleManager  UserRole = "manager"
	RoleEmployee UserRole = "employee"
)

// ======================
// USER ENTITY (DB)
// ======================
type User struct {
	ID       uint `gorm:"primaryKey" json:"id"`
	TenantID uint `gorm:"index;not null" json:"tenant_id" example:"1"`

	Name     string `gorm:"type:varchar(100);not null" json:"name" example:"Budi Santoso"`
	Email    string `gorm:"type:varchar(100);unique;not null" json:"email" example:"budi@company.com"`
	Password string `gorm:"type:varchar(255);not null" json:"-"`

	Role UserRole `gorm:"type:varchar(50);default:employee" json:"role" example:"employee"`

	Attendances []Attendance `gorm:"foreignKey:UserID" json:"attendances,omitempty"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ======================
// FILTER (QUERY)
// ======================
type UserFilter struct {
	Name     string
	Email    string
	Role     UserRole
	TenantID uint

	OrderBy string // created_at, name, email
	Sort    string // asc / desc

	Limit  int
	Offset int
}

// ======================
// REQUEST DTO
// ======================
type CreateUserRequest struct {
	Name     string   `json:"name" binding:"required" example:"Budi Santoso"`
	Email    string   `json:"email" binding:"required,email" example:"budi@company.com"`
	Password string   `json:"password" binding:"required,min=6" example:"123456"`
	Role     UserRole `json:"role" example:"employee"`
}

// ======================
// RESPONSE DTO
// ======================
type UserResponse struct {
	ID        uint      `json:"id" example:"1"`
	Name      string    `json:"name" example:"Budi Santoso"`
	Email     string    `json:"email" example:"budi@company.com"`
	Role      UserRole  `json:"role" example:"employee"`
	TenantID  uint      `json:"tenant_id" example:"1"`
	CreatedAt time.Time `json:"created_at"`
}
