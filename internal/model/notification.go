package model

import (
	"time"

	"gorm.io/gorm"
)

type NotificationType string

const (
	NotificationTypeLeave        NotificationType = "leave"
	NotificationTypeOvertime     NotificationType = "overtime"
	NotificationTypeExpense      NotificationType = "expense"
	NotificationTypePayroll      NotificationType = "payroll"
	NotificationTypeProfile      NotificationType = "profile"
	NotificationTypeSupport      NotificationType = "support"
	NotificationTypeSystem       NotificationType = "system"
	NotificationTypeSubscription NotificationType = "subscription"
)

type Notification struct {
	ID        uint             `gorm:"primaryKey" json:"id"`
	TenantID  uint             `gorm:"index;not null" json:"tenant_id"`
	UserID    uint             `gorm:"index;not null" json:"user_id"`
	Title     string           `gorm:"type:varchar(255);not null" json:"title"`
	Message   string           `gorm:"type:text;not null" json:"message"`
	Type      NotificationType `gorm:"type:varchar(50);not null" json:"type"`
	IsRead    bool             `gorm:"default:false" json:"is_read"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	DeletedAt gorm.DeletedAt   `gorm:"index" json:"-"`
}

type NotificationResponse struct {
	ID        uint             `json:"id"`
	Title     string           `json:"title"`
	Message   string           `json:"message"`
	Type      NotificationType `json:"type"`
	IsRead    bool             `json:"is_read"`
	CreatedAt time.Time        `json:"created_at"`
}

// SSEEvent is the envelope for all Server-Sent Events on the notification stream.
// Type determines what the FE should do: render toast, update badge, or confirm connection.
type SSEEvent struct {
	Type        string      `json:"type"`                // "connected" | "notification" | "unread_count" | "RELOAD_NAV"
	UnreadCount *int64      `json:"unread_count,omitempty"` // present for badge sync
	Data        interface{} `json:"data,omitempty"`     // payload for the event
	EventID     string      `json:"event_id,omitempty"` // notification ID as string, for Last-Event-ID
	Timestamp   int64       `json:"timestamp"`           // unix seconds, Asia/Jakarta
}
