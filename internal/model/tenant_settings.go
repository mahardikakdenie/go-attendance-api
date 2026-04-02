package model

import "time"

type TenantSetting struct {
	ID       uint `gorm:"primaryKey"`
	TenantID uint `gorm:"uniqueIndex" example:"1"`

	OfficeLatitude  float64 `example:"-6.1339179"`
	OfficeLongitude float64 `example:"106.8329504"`
	MaxRadiusMeter  float64 `example:"100"`

	AllowRemote     bool `example:"false"`
	RequireLocation bool `example:"true"`

	ClockInStartTime string `example:"07:00"`
	ClockInEndTime   string `example:"09:00"`

	LateAfterMinute int `example:"480"`

	ClockOutStartTime string `example:"16:00"`
	ClockOutEndTime   string `example:"23:00"`

	RequireSelfie bool `example:"true"`

	AllowMultipleCheck bool `example:"false"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
