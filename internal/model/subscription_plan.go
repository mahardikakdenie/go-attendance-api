package model

import "time"

type SubscriptionPlan struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(50);not null;unique" json:"name"` // Trial, Starter, Business, Enterprise
	MaxEmployees int       `gorm:"default:0" json:"max_employees"`               // 0 means unlimited
	Features     []string  `gorm:"serializer:json;type:jsonb" json:"features"`   // List of allowed modules
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
