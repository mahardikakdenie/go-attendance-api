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

type WorkShiftResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	StartTime string    `json:"startTime"`
	EndTime   string    `json:"endTime"`
	Type      ShiftType `json:"type"`
	Color     string    `json:"color"`
	IsDefault bool      `json:"isDefault"`
}

func (s *WorkShift) BeforeCreate(tx *gorm.DB) (err error) {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return
}

type EventType string

const (
	EventTypeNationalHoliday  EventType = "National Holiday"
	EventTypeCompanyEvent     EventType = "Company Event"
	EventTypeMandatoryLeave   EventType = "Mandatory Leave"
	EventTypeMeeting          EventType = "Meeting"
	EventTypeOther            EventType = "Other"
)

type EventCategory string

const (
	EventCategoryOfficeClosed EventCategory = "OFFICE_CLOSED"
	EventCategoryInformation  EventCategory = "INFORMATION"
)

type CalendarEvent struct {
	ID          uuid.UUID     `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TenantID    uint          `gorm:"not null;index:idx_tenant_event_date" json:"tenant_id"`
	Date        time.Time     `gorm:"type:date;not null;index:idx_tenant_event_date" json:"date"`
	Name        string        `gorm:"type:varchar(255);not null" json:"name"`
	Type        EventType     `gorm:"type:varchar(50);not null" json:"type"`
	Category    EventCategory `gorm:"type:varchar(50);not null;default:'OFFICE_CLOSED'" json:"category"`
	IsPaid      bool          `gorm:"default:true" json:"is_paid"`
	Description string        `gorm:"type:text" json:"description"`
	IsAllUsers  bool          `gorm:"default:true" json:"is_all_users"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`

	Users []User `gorm:"many2many:calendar_event_users;" json:"users,omitempty"`
}

// Keeping old names as aliases for GORM compatibility if needed, 
// but preferred way is to migrate the table.
type Holiday = CalendarEvent
type HolidayType = EventType

func (e *CalendarEvent) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return
}

type EmployeeRoster struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	TenantID  uint      `gorm:"not null" json:"tenant_id"`
	UserID    uint      `gorm:"not null" json:"user_id"`
	Date      time.Time `gorm:"type:date;not null" json:"date"`
	ShiftID   *uuid.UUID `gorm:"type:uuid" json:"shift_id"` // null means fallback to default/global company shift
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
