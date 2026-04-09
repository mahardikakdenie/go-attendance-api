package model

import (
	"time"

	"gorm.io/gorm"
)

type LeaveStatus string

const (
	LeaveStatusPending  LeaveStatus = "pending"
	LeaveStatusApproved LeaveStatus = "approved"
	LeaveStatusRejected LeaveStatus = "rejected"
)

type LeaveType struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	TenantID  uint           `gorm:"index;not null" json:"tenant_id"`
	Name      string         `gorm:"type:varchar(100);not null" json:"name"` // e.g., Annual, Sick, Paternity
	Code      string         `gorm:"type:varchar(20);not null" json:"code"` // e.g., ANNUAL, SICK
	DefaultDays int          `json:"default_days"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type LeaveBalance struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	LeaveTypeID uint      `gorm:"index;not null" json:"leave_type_id"`
	Balance     int       `gorm:"not null" json:"balance"` // Remaining days
	Year        int       `gorm:"not null" json:"year"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	LeaveType *LeaveType `gorm:"foreignKey:LeaveTypeID" json:"leave_type,omitempty"`
}

type Leave struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	TenantID    uint           `gorm:"index;not null" json:"tenant_id"`
	UserID      uint           `gorm:"index;not null" json:"user_id"`
	LeaveTypeID uint           `gorm:"index;not null" json:"leave_type_id"`
	StartDate   time.Time      `gorm:"not null" json:"start_date"`
	EndDate     time.Time      `gorm:"not null" json:"end_date"`
	Reason      string         `gorm:"type:text" json:"reason"`
	Status      LeaveStatus    `gorm:"type:varchar(20);default:'pending'" json:"status"`
	AdminNotes  string         `gorm:"type:text" json:"admin_notes,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	User      *User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	LeaveType *LeaveType `gorm:"foreignKey:LeaveTypeID" json:"leave_type,omitempty"`
}

type LeaveRequest struct {
	LeaveTypeID uint   `json:"leave_type_id" binding:"required"`
	StartDate   string `json:"start_date" binding:"required" example:"2026-05-01"`
	EndDate     string `json:"end_date" binding:"required" example:"2026-05-03"`
	Reason      string `json:"reason" binding:"required"`
}

type LeaveResponse struct {
	ID          uint        `json:"id"`
	UserID      uint        `json:"user_id"`
	LeaveTypeID uint        `json:"leave_type_id"`
	LeaveType   string      `json:"leave_type_name"`
	StartDate   time.Time   `json:"start_date"`
	EndDate     time.Time   `json:"end_date"`
	TotalDays   int         `json:"total_days"`
	Reason      string      `json:"reason"`
	Status      LeaveStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
}
