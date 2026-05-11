package model

import (
	"time"

	"gorm.io/gorm"
)

type Menu struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	ParentID     *uint          `gorm:"index" json:"parent_id"`
	Key          string         `gorm:"type:varchar(100);unique;not null" json:"key"`
	Label        string         `gorm:"type:varchar(100);not null" json:"label"`
	Icon         string         `gorm:"type:varchar(100)" json:"icon"`
	Path         string         `gorm:"type:varchar(255)" json:"path"`
	AllowedRoles []string       `gorm:"serializer:json;type:jsonb" json:"allowed_roles"`
	Permission   string         `gorm:"type:varchar(100)" json:"permission"`
	Module       string         `gorm:"type:varchar(50)" json:"module"`
	IsSystem     bool           `gorm:"default:false" json:"is_system"`
	SortOrder    int            `gorm:"default:0" json:"sort_order"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`

	Children []Menu `gorm:"foreignKey:ParentID" json:"children,omitempty"`
}

type MenuResponse struct {
	ID         uint           `json:"id"`
	Key        string         `json:"key"`
	Label      string         `json:"label"`
	Icon       string         `json:"icon"`
	Path       string         `json:"path,omitempty"`
	Module     string         `json:"module,omitempty"`
	Permission string         `json:"permission,omitempty"`
	Children   []MenuResponse `json:"children,omitempty"`
}
