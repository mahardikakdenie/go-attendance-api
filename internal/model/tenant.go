package model

import "time"

type Tenant struct {
	ID              uint           `gorm:"primaryKey" json:"id"`
	Name            string         `gorm:"not null" json:"name"`
	Code            string         `gorm:"uniqueIndex" json:"code"`
	IsSuspended     bool           `gorm:"default:false" json:"is_suspended"`
	SuspendedReason string         `gorm:"type:text" json:"suspended_reason,omitempty"`
	TenantSettings  *TenantSetting `gorm:"foreignKey:TenantID" json:"tenant_settings,omitempty"`
	Subscription    *Subscription  `gorm:"foreignKey:TenantID" json:"subscription,omitempty"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`
}

type TenantResponse struct {
	ID              uint           `json:"id"`
	Name            string         `json:"name"`
	Plan            string         `json:"plan"` // This will be populated from Subscription relation in service layer
	IsSuspended     bool           `json:"is_suspended"`
	SuspendedReason string         `json:"suspended_reason,omitempty"`
	TenantSettings  *TenantSetting `json:"tenant_settings,omitempty"`
	Subscription    *Subscription  `json:"subscription,omitempty"`
}
