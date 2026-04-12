package model

import "time"

type BaseRole string

const (
	BaseRoleSuperAdmin BaseRole = "SUPERADMIN"
	BaseRoleAdmin      BaseRole = "ADMIN"
	BaseRoleHR         BaseRole = "HR"
	BaseRoleFinance    BaseRole = "FINANCE"
	BaseRoleEmployee   BaseRole = "EMPLOYEE"
)

type Role struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	TenantID    *uint     `gorm:"index" json:"tenant_id"`
	Name        string    `gorm:"type:varchar(50);not null" json:"name" example:"admin"`
	Description string    `gorm:"type:text" json:"description"`
	BaseRole    BaseRole  `gorm:"type:varchar(20);not null;default:'EMPLOYEE'" json:"base_role"`
	IsSystem    bool      `gorm:"default:false" json:"is_system"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Permissions []Permission `gorm:"many2many:role_permissions;" json:"permissions,omitempty"`
}

type RoleResponse struct {
	ID          uint         `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	BaseRole    BaseRole     `json:"base_role"`
	IsSystem    bool         `json:"is_system"`
	Permissions []Permission `json:"permissions,omitempty"`
}
