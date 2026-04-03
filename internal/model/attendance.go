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

	ClockInTime  time.Time  `json:"clock_in_time"`
	ClockOutTime *time.Time `json:"clock_out_time,omitempty"`

	ClockInLatitude   float64  `json:"clock_in_latitude"`
	ClockInLongitude  float64  `json:"clock_in_longitude"`
	ClockOutLatitude  *float64 `json:"clock_out_latitude,omitempty"`
	ClockOutLongitude *float64 `json:"clock_out_longitude,omitempty"`

	ClockInMediaUrl  string  `json:"clock_in_media_url"`
	ClockOutMediaUrl *string `json:"clock_out_media_url,omitempty"`

	Status AttendanceStatus `json:"status"`
}

type AttendanceFilter struct {
	UserID   uint
	Status   AttendanceStatus
	limit    int
	offset   int
	includes []string
	DateFrom *time.Time
	DateTo   *time.Time
}

var ErrNotFound = errors.New("data not found")
