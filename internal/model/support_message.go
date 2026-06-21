package model

import (
	"time"

	"github.com/google/uuid"
)

type SupportCategory string

const (
	SupportCategoryTechnical   SupportCategory = "TECHNICAL"
	SupportCategoryBilling     SupportCategory = "BILLING"
	SupportCategoryFeature     SupportCategory = "FEATURE"
	SupportCategoryAccount     SupportCategory = "ACCOUNT"
	SupportCategoryIntegration SupportCategory = "INTEGRATION"
	SupportCategoryOther       SupportCategory = "OTHER"
)

type SupportStatus string

const (
	SupportStatusPending    SupportStatus = "PENDING"
	SupportStatusInProgress SupportStatus = "IN_PROGRESS"
	SupportStatusResolved   SupportStatus = "RESOLVED"
	SupportStatusClosed     SupportStatus = "CLOSED"
)

type SupportPriority string

const (
	SupportPriorityLow      SupportPriority = "LOW"
	SupportPriorityMedium   SupportPriority = "MEDIUM"
	SupportPriorityHigh     SupportPriority = "HIGH"
	SupportPriorityUrgent   SupportPriority = "URGENT"
)

type SupportMessage struct {
	ID            uuid.UUID        `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TenantID      uint             `gorm:"not null;index" json:"tenant_id"`
	UserID        uint             `gorm:"not null;index" json:"user_id"`
	Subject       string           `gorm:"type:varchar(255);not null;index" json:"subject"`
	Message       string           `gorm:"type:text;not null" json:"message"`
	Category      SupportCategory  `gorm:"type:varchar(20);not null;index" json:"category"`
	Priority      SupportPriority  `gorm:"type:varchar(20);not null;default:'MEDIUM';index" json:"priority"`
	Status        SupportStatus    `gorm:"type:varchar(20);default:'PENDING';index" json:"status"`
	IsRead        bool             `gorm:"default:false;index" json:"is_read"`
	AssignedToID  *uint            `gorm:"index" json:"assigned_to_id,omitempty"`
	AttachmentURL string           `gorm:"type:varchar(500)" json:"attachment_url,omitempty"`
	CreatedAt     time.Time        `json:"created_at"`

	Tenant     Tenant         `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`
	User       User           `gorm:"foreignKey:UserID" json:"user,omitempty"`
	AssignedTo *User          `gorm:"foreignKey:AssignedToID" json:"assigned_to,omitempty"`
	Replies    []SupportReply `gorm:"foreignKey:MessageID" json:"replies,omitempty"`
}

type SupportMessageFilter struct {
	Search   string
	Category SupportCategory
	Status   SupportStatus
	Priority SupportPriority
	Limit    int
	Offset   int
}
