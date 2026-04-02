package model

import "time"

type Tenant struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Code      string `gorm:"uniqueIndex"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
