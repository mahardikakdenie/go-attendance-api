package model

import (
	"encoding/json"
	"time"
)

type UserSetting struct {
	ID                       uint             `gorm:"primaryKey" json:"id"`
	UserID                   uint             `gorm:"not null;uniqueIndex" json:"user_id" example:"1"`
	TenantID                 uint             `gorm:"not null;index" json:"tenant_id" example:"1"`
	
	OfficeLatitude           *float64         `json:"office_latitude,omitempty"`
	OfficeLongitude          *float64         `json:"office_longitude,omitempty"`
	MaxRadiusMeter           *float64         `json:"max_radius_meter,omitempty"`
	AllowRemote              *bool            `json:"allow_remote,omitempty"`
	RequireLocation          *bool            `json:"require_location,omitempty"`
	ClockInStartTime         *string          `json:"clock_in_start_time,omitempty"`
	ClockInEndTime           *string          `json:"clock_in_end_time,omitempty"`
	LateAfterMinute          *int             `json:"late_after_minute,omitempty"`
	ClockOutStartTime        *string          `json:"clock_out_start_time,omitempty"`
	ClockOutEndTime          *string          `json:"clock_out_end_time,omitempty"`
	RequireSelfie            *bool            `json:"require_selfie,omitempty"`
	AllowMultipleCheck       *bool            `json:"allow_multiple_check,omitempty"`
	AttendanceSessionsConfig *json.RawMessage `gorm:"type:jsonb" json:"attendance_sessions_config,omitempty"`

	CreatedAt                time.Time        `json:"created_at"`
	UpdatedAt                time.Time        `json:"updated_at"`

	User   *User   `gorm:"foreignKey:UserID" json:"-"`
	Tenant *Tenant `gorm:"foreignKey:TenantID" json:"-"`
}
