package model

import (
	"time"

	"gorm.io/gorm"
)

type SubscriptionFeature struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	FeatureKey  string         `gorm:"type:varchar(50);unique;not null" json:"feature_key"`
	Label       string         `gorm:"type:varchar(100);not null" json:"label"`
	Description string         `gorm:"type:text" json:"description"`
	IsActive    bool           `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
