package model

import "time"

type UserRole string

const (
	RoleSuperAdmin UserRole = "superadmin"
	RoleAdmin      UserRole = "admin"
	RoleHR         UserRole = "hr"
	RoleEmployee   UserRole = "employee"
)

type User struct {
	ID       uint `gorm:"primaryKey" json:"id" example:"1"`
	TenantID uint `gorm:"index;not null" json:"tenant_id" example:"1"`

	Tenant *Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`

	Name        string `gorm:"type:varchar(100);not null" json:"name" example:"Budi Santoso" binding:"required,min=3,max=100"`
	Email       string `gorm:"type:varchar(100);unique;not null" json:"email" example:"budi@company.com" binding:"required,email"`
	Password    string `gorm:"type:varchar(255);not null" json:"-"` // tetap hidden
	EmployeeID  string `gorm:"type:varchar(50);uniqueIndex" json:"employee_id" example:"FT-001"`
	Department  string `gorm:"type:varchar(100)" json:"department" example:"IT"`
	Address     string `gorm:"type:text" json:"address" example:"Jl. Sudirman No. 1"`
	PhoneNumber string `gorm:"type:varchar(20)" json:"phone_number" example:"08123456789"`

	RoleID uint  `gorm:"not null" json:"role_id" example:"1"`
	Role   *Role `gorm:"foreignKey:RoleID" json:"role,omitempty"`

	Attendances      []Attendance     `gorm:"foreignKey:UserID" json:"attendances,omitempty"`
	RecentActivities []RecentActivity `gorm:"foreignKey:UserID" json:"recent_activities,omitempty"`

	MediaUrl string `gorm:"type:varchar(255)" json:"media_url" example:"https://cdn.example.com/profile/budi.jpg" binding:"omitempty,url"`

	CreatedAt time.Time `json:"created_at" example:"2026-04-05T12:00:00Z"`
	UpdatedAt time.Time `json:"updated_at" example:"2026-04-05T12:30:00Z"`
}

type UserFilter struct {
	Name       string
	Email      string
	RoleID     uint
	TenantID   uint
	EmployeeID string

	OrderBy string
	Sort    string

	Limit  int
	Offset int
}

type CreateUserRequest struct {
	Name        string `json:"name" binding:"required" example:"Budi Santoso"`
	Email       string `json:"email" binding:"required,email" example:"budi@company.com"`
	Password    string `json:"password" binding:"required,min=6" example:"123456"`
	RoleID      uint   `json:"role_id" example:"1"`
	TenantID    uint   `json:"tenant_id" example:"1"`
	Department  string `json:"department" example:"IT"`
	Address     string `json:"address" example:"Jl. Sudirman No. 1"`
	PhoneNumber string `json:"phone_number" example:"08123456789"`
}

type UserResponse struct {
	ID          uint          `json:"id" example:"1"`
	Name        string        `json:"name" example:"Budi Santoso"`
	Email       string        `json:"email" example:"budi@company.com"`
	Role        *RoleResponse `json:"role,omitempty"`
	TenantID    uint          `json:"tenant_id" example:"1"`
	EmployeeID  string        `json:"employee_id" example:"FT-001"`
	Department  string        `json:"department" example:"IT"`
	Address     string        `json:"address" example:"Jl. Sudirman No. 1"`
	MediaUrl    string        `gorm:"type:varchar(255)" json:"media_url" example:"https://cdn.example.com/profile/budi.jpg" binding:"omitempty,url"`
	PhoneNumber string        `json:"phone_number" example:"08123456789"`
	CreatedAt   time.Time     `json:"created_at"`

	Tenant           *TenantResponse          `json:"tenant,omitempty"`
	Attendances      []AttendanceResponse     `json:"attendances,omitempty"`
	RecentActivities []RecentActivityResponse `json:"recent_activities,omitempty"`
}
