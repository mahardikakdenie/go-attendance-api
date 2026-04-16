package model

import (
	"time"

	"github.com/google/uuid"
)

type SupportCategory string

const (
	SupportCategoryTechnical SupportCategory = "TECHNICAL"
	SupportCategoryBilling   SupportCategory = "BILLING"
	SupportCategoryFeature   SupportCategory = "FEATURE"
	SupportCategoryOther     SupportCategory = "OTHER"
)

type SupportStatus string

const (
	SupportStatusPending    SupportStatus = "PENDING"
	SupportStatusInProgress SupportStatus = "IN_PROGRESS"
	SupportStatusResolved   SupportStatus = "RESOLVED"
)

type SupportMessage struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TenantID  uint            `gorm:"not null" json:"tenant_id"`
	UserID    uint            `gorm:"not null" json:"user_id"`
	Subject   string          `gorm:"type:varchar(255);not null" json:"subject"`
	Message   string          `gorm:"type:text;not null" json:"message"`
	Category  SupportCategory `gorm:"type:varchar(20);not null" json:"category"`
	Status    SupportStatus   `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	CreatedAt time.Time       `json:"created_at"`

	Tenant Tenant `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	User   User   `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
