package model

import "time"

type Tenant struct {
	ID             uint           `gorm:"primaryKey"`
	Name           string         `gorm:"not null"`
	Code           string         `gorm:"uniqueIndex"`
	TenantSettings *TenantSetting `gorm:"foreignKey:TenantID" json:"tenant_settings,omitempty"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type TenantResponse struct {
	ID             uint           `json:"id"`
	Name           string         `json:"name"`
	TenantSettings *TenantSetting `json:"tenant_settings,omitempty"`
}
