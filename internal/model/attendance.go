package model

import (
	"time"

	"github.com/google/uuid"
)

type AttendanceRequest struct {
	EmployeeID int     `json:"employee_id" binding:"required"`
	Action     string  `json:"action" binding:"required"`
	Latitude   float64 `json:"latitude" binding:"required"`
	Longitude  float64 `json:"longitude" binding:"required"`
	MediaUrl   string  `json:"media_url"`
}

type AttendanceResponse struct {
	ID                uuid.UUID  `json:"id"`
	EmployeeID        int        `json:"employee_id"`
	ClockInTime       time.Time  `json:"clock_in_time"`
	ClockOutTime      *time.Time `json:"clock_out_time,omitempty"`
	ClockInLatitude   float64    `json:"clock_in_latitude"`
	ClockInLongitude  float64    `json:"clock_in_longitude"`
	ClockOutLatitude  *float64   `json:"clock_out_latitude,omitempty"`
	ClockOutLongitude *float64   `json:"clock_out_longitude,omitempty"`
	Status            string     `json:"status"`
	MediaUrl          string     `json:"media_url"`
}

type Attendance struct {
	ID     uuid.UUID `gorm:"type:uuid;default:gen_random_uuid();primaryKey" json:"id"`
	UserID uint      `gorm:"not null" json:"user_id"`
	User   User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"user,omitempty"`

	ClockInTime      time.Time `gorm:"not null" json:"clock_in_time"`
	ClockInLatitude  float64   `gorm:"not null" json:"clock_in_latitude"`
	ClockInLongitude float64   `gorm:"not null" json:"clock_in_longitude"`

	ClockOutTime      *time.Time `json:"clock_out_time"`
	ClockOutLatitude  *float64   `json:"clock_out_latitude"`
	ClockOutLongitude *float64   `json:"clock_out_longitude"`

	Status string `gorm:"type:varchar(50)" json:"status"`

	MediaUrl string `gorm:"type:varchar(255)" json:"media_url"`
}
