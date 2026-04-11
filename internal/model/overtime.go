package model

import (
	"time"
)

type OvertimeStatus string

const (
	OvertimeStatusPending  OvertimeStatus = "pending"
	OvertimeStatusApproved OvertimeStatus = "approved"
	OvertimeStatusRejected OvertimeStatus = "rejected"
)

type Overtime struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"index;not null" json:"user_id"`
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	TenantID  uint      `gorm:"index;not null" json:"tenant_id"`
	Tenant    *Tenant   `gorm:"foreignKey:TenantID" json:"tenant,omitempty"`

	Date      time.Time `gorm:"type:date;not null" json:"date"`
	StartTime string    `gorm:"type:varchar(5);not null" json:"start_time" example:"17:00"`
	EndTime   string    `gorm:"type:varchar(5);not null" json:"end_time" example:"19:00"`
	Reason    string    `gorm:"type:text" json:"reason"`

	Status     OvertimeStatus `gorm:"type:varchar(20);default:'pending'" json:"status"`
	AdminNotes string         `gorm:"type:text" json:"admin_notes"`

	ApprovedBy *uint      `json:"approved_by"`
	ApprovedAt *time.Time `json:"approved_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateOvertimeRequest struct {
	Date      string `json:"date" binding:"required" example:"2026-04-08"`
	StartTime string `json:"start_time" binding:"required" example:"17:00"`
	EndTime   string `json:"end_time" binding:"required" example:"19:00"`
	Reason    string `json:"reason" binding:"required"`
}

type ApproveOvertimeRequest struct {
	AdminNotes string `json:"admin_notes"`
}

type OvertimeResponse struct {
	ID          uint           `json:"id"`
	UserID      uint           `json:"user_id"`
	User        *UserResponse  `json:"user,omitempty"`
	Date        time.Time      `json:"date"`
	StartTime   string         `json:"start_time"`
	EndTime     string         `json:"end_time"`
	Reason      string         `json:"reason"`
	Status      OvertimeStatus `json:"status"`
	AdminNotes  string         `json:"admin_notes"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

type OvertimeFilter struct {
	UserID         uint
	TenantID       uint
	Status         OvertimeStatus
	AllowedRoleIDs []uint
	DateFrom       *time.Time
	DateTo         *time.Time
	Limit          int
	Offset         int
}
