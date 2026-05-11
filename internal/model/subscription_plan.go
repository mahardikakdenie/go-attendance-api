package model

import "time"

type SubscriptionPlan struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	Name         string    `gorm:"type:varchar(50);not null;unique" json:"name"` // Trial, Starter, Business, Enterprise
	Price        float64   `gorm:"type:decimal(15,2);default:0" json:"price"`
	Days         int       `gorm:"default:0" json:"days"`                      // Duration in days (e.g., 30 for monthly)
	MaxEmployees int       `gorm:"default:0" json:"max_employees"`             // 0 means unlimited
	Features     []string  `gorm:"serializer:json;type:jsonb" json:"features"` // List of allowed modules
	IsActive     bool      `gorm:"default:true" json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
