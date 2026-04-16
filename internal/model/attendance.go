package model

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type AttendanceAction string

const (
	ClockIn  AttendanceAction = "clock_in"
	ClockOut AttendanceAction = "clock_out"
)

type AttendanceStatus string

const (
	StatusWorking AttendanceStatus = "working"
	StatusDone    AttendanceStatus = "done"
	StatusLate    AttendanceStatus = "late"
)

type AttendanceRequest struct {
	Action    AttendanceAction `json:"action" example:"clock_in" binding:"required,oneof=clock_in clock_out"`
	Latitude  float64          `json:"latitude" example:"-6.1339179" binding:"required"`
	Longitude float64          `json:"longitude" example:"106.8329504" binding:"required"`
	MediaUrl  string           `json:"media_url" example:"https://i.pinimg.com/control1/736x/41/e7/99/41e799436291fdfb4c969b80913a19fe.jpg"`
}
type Attendance struct {
	ID       uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	UserID   uint      `gorm:"not null;index"`
	User     User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	TenantID uint      `gorm:"index"`
	Tenant   Tenant    `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`

	ClockInTime      time.Time `gorm:"not null;index"`
	ClockInLatitude  float64   `gorm:"not null"`
	ClockInLongitude float64   `gorm:"not null"`
	ClockInMediaUrl  string    `gorm:"type:varchar(255)"`

	ClockOutTime      *time.Time
	ClockOutLatitude  *float64
	ClockOutLongitude *float64
	ClockOutMediaUrl  *string

	Status AttendanceStatus `gorm:"type:varchar(50)"`
}

type AttendanceResponse struct {
	ID     uuid.UUID `json:"id"`
	UserID uint      `json:"user_id"`
	User   *UserResponse `json:"user,omitempty"`

	ClockInTime  time.Time  `json:"clock_in_time"`
	ClockOutTime *time.Time `json:"clock_out_time,omitempty"`

	ClockInLatitude   float64  `json:"clock_in_latitude"`
	ClockInLongitude  float64  `json:"clock_in_longitude"`
	ClockOutLatitude  *float64 `json:"clock_out_latitude,omitempty"`
	ClockOutLongitude *float64 `json:"clock_out_longitude,omitempty"`

	ClockInMediaUrl  string  `json:"clock_in_media_url"`
	ClockOutMediaUrl *string `json:"clock_out_media_url,omitempty"`

	Status    AttendanceStatus `json:"status"`
	CreatedAt time.Time        `json:"created_at"`
}

type AttendanceSummaryResponse struct {
	TotalRecord          int64   `json:"total_record"`
	TotalRecordDiff      float64 `json:"total_record_diff"`

	OntimeSummary          int64   `json:"ontime_summary"`
	OntimeSummaryDiff      float64 `json:"ontime_summary_diff"`

	LateSummary          int64   `json:"late_summary"`
	LateSummaryDiff      float64 `json:"late_summary_diff"`
}

type AttendanceFilter struct {
	UserID         uint
	TenantID       uint
	Search         string
	Status         AttendanceStatus
	AllowedRoleIDs []uint
	limit          int
	offset         int
	includes       []string
	DateFrom       *time.Time
	DateTo         *time.Time
}

var ErrNotFound = errors.New("data not found")

type CorrectionStatus string

const (
	CorrectionPending  CorrectionStatus = "PENDING"
	CorrectionApproved CorrectionStatus = "APPROVED"
	CorrectionRejected CorrectionStatus = "REJECTED"
)

type AttendanceCorrection struct {
	ID           uint             `gorm:"primaryKey" json:"id"`
	TenantID     uint             `gorm:"not null;index" json:"tenant_id"`
	UserID       uint             `gorm:"not null;index" json:"user_id"`
	AttendanceID *uuid.UUID       `gorm:"type:uuid" json:"attendance_id"` // Null if user completely forgot to clock in and out
	Date         time.Time        `gorm:"type:date;not null" json:"date"`
	ClockInTime  *time.Time       `json:"clock_in_time"`
	ClockOutTime *time.Time       `json:"clock_out_time"`
	Reason       string           `gorm:"type:text;not null" json:"reason"`
	Status       CorrectionStatus `gorm:"type:varchar(20);default:'PENDING'" json:"status"`
	ApprovedBy   *uint            `json:"approved_by"`
	ApprovedAt   *time.Time       `json:"approved_at"`
	AdminNotes   string           `gorm:"type:text" json:"admin_notes"`
	CreatedAt    time.Time        `json:"created_at"`
	UpdatedAt    time.Time        `json:"updated_at"`

	User       *User       `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Attendance *Attendance `gorm:"foreignKey:AttendanceID" json:"attendance,omitempty"`
	Approver   *User       `gorm:"foreignKey:ApprovedBy" json:"approver,omitempty"`
}

// Request DTOs
type CreateCorrectionRequest struct {
	AttendanceID *uuid.UUID `json:"attendance_id"`
	Date         string     `json:"date" binding:"required"` // YYYY-MM-DD
	ClockInTime  *string    `json:"clock_in_time"`           // HH:mm:ss
	ClockOutTime *string    `json:"clock_out_time"`          // HH:mm:ss
	Reason       string     `json:"reason" binding:"required"`
}

type ReviewCorrectionRequest struct {
	AdminNotes string `json:"admin_notes"`
}

type AttendanceCorrectionResponse struct {
	ID           uint             `json:"id"`
	UserID       uint             `json:"user_id"`
	UserName     string           `json:"user_name"`
	Date         string           `json:"date"`
	ClockInTime  *string          `json:"clock_in_time"`
	ClockOutTime *string          `json:"clock_out_time"`
	Reason       string           `json:"reason"`
	Status       CorrectionStatus `json:"status"`
	AdminNotes   string           `json:"admin_notes"`
	CreatedAt    time.Time        `json:"created_at"`
}
