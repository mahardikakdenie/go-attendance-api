package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type RoleHierarchy struct {
	ID           uuid.UUID `gorm:"primaryKey;type:uuid" json:"id"`
	TenantID     uint      `gorm:"index;not null" json:"tenant_id"`
	ParentRoleID uint      `gorm:"index;not null" json:"parent_role_id"`
	ChildRoleID  uint      `gorm:"index;not null" json:"child_role_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	ParentRole *Role `gorm:"foreignKey:ParentRoleID" json:"parent_role,omitempty"`
	ChildRole  *Role `gorm:"foreignKey:ChildRoleID" json:"child_role,omitempty"`
}

func (rh *RoleHierarchy) BeforeCreate(tx *gorm.DB) (err error) {
	rh.ID = uuid.New()
	return
}
