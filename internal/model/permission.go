package model

import "time"

type Permission struct {
	ID        string    `gorm:"primaryKey;type:varchar(100)" json:"id" example:"attendance.view"`
	Module    string    `gorm:"type:varchar(50);not null" json:"module" example:"attendance"`
	Action    string    `gorm:"type:varchar(50);not null" json:"action" example:"view"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type RolePermission struct {
	RoleID       uint   `gorm:"primaryKey"`
	PermissionID string `gorm:"primaryKey;type:varchar(100)"`
}
