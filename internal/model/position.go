package model

import "time"

type Position struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TenantID  uint      `gorm:"index;not null" json:"tenant_id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`  // e.g., CEO, VP, Manager
	Level     int       `gorm:"not null" json:"level"`                   // 1 for C-Level, 2 for VP, etc.
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrgNode struct {
	ID         uint       `json:"id"`
	Name       string     `json:"name"`
	Position   string     `json:"position"`
	Level      int        `json:"level"`
	Avatar     string     `json:"avatar"`
	Subordinates []OrgNode `json:"subordinates,omitempty"`
}
