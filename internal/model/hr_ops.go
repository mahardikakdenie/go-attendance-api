package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ShiftType string

const (
	ShiftTypeMorning   ShiftType = "Morning"
	ShiftTypeAfternoon ShiftType = "Afternoon"
	ShiftTypeNight     ShiftType = "Night"
)

type WorkShift struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TenantID  uint      `gorm:"not null" json:"tenant_id"`
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`
	StartTime string    `gorm:"type:varchar(5);not null" json:"startTime"` // Format: HH:mm
	EndTime   string    `gorm:"type:varchar(5);not null" json:"endTime"`   // Format: HH:mm
	Type      ShiftType `gorm:"type:varchar(20);not null" json:"type"`
	Color     string    `gorm:"type:varchar(20)" json:"color"` // e.g., bg-emerald-500
	IsDefault bool      `gorm:"default:false" json:"isDefault"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (s *WorkShift) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

type HolidayType string

const (
	HolidayTypeNational       HolidayType = "National Holiday"
	HolidayTypeCompanyEvent   HolidayType = "Company Event"
	HolidayTypeMandatoryLeave HolidayType = "Mandatory Leave"
	HolidayTypeOther          HolidayType = "Other"
)

type Holiday struct {
	ID        uuid.UUID   `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TenantID  uint        `gorm:"not null;uniqueIndex:idx_tenant_holiday_date" json:"tenant_id"`
	Date      time.Time   `gorm:"type:date;not null;uniqueIndex:idx_tenant_holiday_date" json:"date"`
	Name      string      `gorm:"type:varchar(255);not null" json:"name"`
	Type      HolidayType `gorm:"type:varchar(50);not null" json:"type"`
	IsPaid    bool        `gorm:"default:true" json:"is_paid"`
	CreatedAt time.Time   `json:"created_at"`
	UpdatedAt time.Time   `json:"updated_at"`
}

func (h *Holiday) BeforeCreate(tx *gorm.DB) (err error) {
	if h.ID == uuid.Nil {
		h.ID = uuid.New()
	}
	return
}

type EmployeeRoster struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TenantID  uint      `gorm:"not null" json:"tenant_id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	Date      time.Time `gorm:"type:date;not null" json:"date"`
	ShiftID   *uuid.UUID `gorm:"type:uuid" json:"shift_id"` // null means "off"
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	User  User       `gorm:"foreignKey:UserID" json:"-"`
	Shift *WorkShift `gorm:"foreignKey:ShiftID" json:"shift,omitempty"`
}

type LifecycleStatus string

const (
	LifecycleStatusOnboarding  LifecycleStatus = "ONBOARDING"
	LifecycleStatusActive      LifecycleStatus = "ACTIVE"
	LifecycleStatusOffboarding LifecycleStatus = "OFFBOARDING"
)

type LifecycleTask struct {
	ID        uuid.UUID       `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TenantID  uint            `gorm:"not null" json:"tenant_id"`
	TaskName  string          `gorm:"type:varchar(255);not null" json:"task_name"`
	Category  LifecycleStatus `gorm:"type:varchar(50);not null" json:"category"`
	IsSystem  bool            `gorm:"default:false" json:"is_system"`
	CreatedAt time.Time       `json:"created_at"`
}

func (t *LifecycleTask) BeforeCreate(tx *gorm.DB) (err error) {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return
}

type EmployeeLifecycleTask struct {
	ID          uint       `gorm:"primaryKey" json:"id"`
	UserID      uint       `gorm:"not null" json:"user_id"`
	TaskID      uuid.UUID  `gorm:"type:uuid;not null" json:"task_id"`
	IsCompleted bool       `gorm:"default:false" json:"is_completed"`
	CompletedAt *time.Time `json:"completed_at"`
	UpdatedAt   time.Time  `json:"updated_at"`

	Task LifecycleTask `gorm:"foreignKey:TaskID" json:"task"`
}
