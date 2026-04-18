package model

import "time"

type Tenant struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"not null" json:"name"`
	Code            string         `gorm:"uniqueIndex" json:"code"`
	Plan            string         `gorm:"type:varchar(50);default:'Basic'" json:"plan"` // Basic, Pro, Enterprise
	IsSuspended     bool           `gorm:"default:false" json:"is_suspended"`
	SuspendedReason string         `gorm:"type:text" json:"suspended_reason,omitempty"`
	TenantSettings  *TenantSetting `gorm:"foreignKey:TenantID" json:"tenant_settings,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type TenantResponse struct {
	ID              uint           `json:"id"`
	Name            string         `json:"name"`
	Plan            string         `json:"plan"`
	IsSuspended     bool           `json:"is_suspended"`
	SuspendedReason string         `json:"suspended_reason,omitempty"`
	TenantSettings  *TenantSetting `json:"tenant_settings,omitempty"`
}

