package model

import "time"

type TenantSetting struct {
	ID                 uint      `gorm:"primaryKey" json:"id" example:"1"`
	TenantID           uint      `gorm:"uniqueIndex" json:"tenant_id" example:"1"`
	Tenant             Tenant    `gorm:"foreignKey:TenantID" json:"tenant"`
	OfficeLatitude     float64   `json:"office_latitude" example:"-6.1339179"`
	OfficeLongitude    float64   `json:"office_longitude" example:"106.8329504"`
	MaxRadiusMeter     float64   `json:"max_radius_meter" example:"100"`
	AllowRemote        bool      `json:"allow_remote" example:"false"`
	RequireLocation    bool      `json:"require_location" example:"true"`
	ClockInStartTime   string    `json:"clock_in_start_time" example:"07:00"`
	ClockInEndTime     string    `json:"clock_in_end_time" example:"09:00"`
	LateAfterMinute    int       `json:"late_after_minute" example:"480"`
	ClockOutStartTime  string    `json:"clock_out_start_time" example:"16:00"`
	ClockOutEndTime    string    `json:"clock_out_end_time" example:"23:00"`
	RequireSelfie      bool      `json:"require_selfie" example:"true"`
	AllowMultipleCheck bool      `json:"allow_multiple_check" example:"false"`
	CreatedAt          time.Time `json:"created_at" example:"2026-04-07T13:21:24Z"`
	UpdatedAt          time.Time `json:"updated_at" example:"2026-04-07T13:21:24Z"`
}
