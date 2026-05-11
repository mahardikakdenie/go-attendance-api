package model

import (
	"time"

	"gorm.io/gorm"
)

type AuditLog struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	UserID    uint           `gorm:"index;not null" json:"user_id"` // Who performed the action
	Action    string         `gorm:"type:varchar(100);not null" json:"action"`
	Entity    string         `gorm:"type:varchar(100)" json:"entity"`
	EntityID  string         `gorm:"type:varchar(100)" json:"entity_id"`
	OldValue  string         `gorm:"type:text" json:"old_value"`
	NewValue  string         `gorm:"type:text" json:"new_value"`
	IPAddress string         `gorm:"type:varchar(50)" json:"ip_address"`
	CreatedAt time.Time      `json:"created_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	User *User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}
