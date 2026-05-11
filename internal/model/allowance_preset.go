package model

import "time"

type AllowanceType string

const (
	AllowanceTypeFixed    AllowanceType = "fixed"
	AllowanceTypeVariable AllowanceType = "variable"
)

type AllowancePreset struct {
	ID        uint          `gorm:"primaryKey" json:"id"`
	Name      string        `gorm:"type:varchar(100);not null" json:"name"`
	Type      AllowanceType `gorm:"type:varchar(20);not null" json:"type"`
	CreatedAt time.Time     `json:"created_at"`
	UpdatedAt time.Time     `json:"updated_at"`
}
