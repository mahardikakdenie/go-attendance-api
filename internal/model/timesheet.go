package model

import (
	"time"

	"github.com/google/uuid"
)

type ProjectStatus string

const (
	ProjectStatusActive    ProjectStatus = "Active"
	ProjectStatusCompleted ProjectStatus = "Completed"
	ProjectStatusOnHold    ProjectStatus = "On Hold"
)

type Project struct {
	ID          uint          `gorm:"primaryKey" json:"id"`
	TenantID    uint          `gorm:"not null;index" json:"tenant_id"`
	Name        string        `gorm:"type:varchar(255);not null" json:"name" binding:"required"`
	Description string        `gorm:"type:text" json:"description"`
	ClientName  string        `gorm:"type:varchar(255)" json:"client_name"`
	StartDate   *time.Time    `json:"start_date"`
	EndDate     *time.Time    `json:"end_date"`
	Status      ProjectStatus `gorm:"type:varchar(20);default:'Active'" json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at"`

	Tasks       []Task           `gorm:"foreignKey:ProjectID" json:"tasks,omitempty"`
	Timesheets  []TimesheetEntry `gorm:"foreignKey:ProjectID" json:"timesheets,omitempty"`
}

type Task struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	ProjectID   uint      `gorm:"not null;index" json:"project_id" binding:"required"`
	UserID      uint      `gorm:"not null;index" json:"user_id"` // Creator
	Name        string    `gorm:"type:varchar(255);not null" json:"name" binding:"required"`
	Description string    `gorm:"type:text" json:"description"`
	IsCompleted bool      `gorm:"default:false" json:"is_completed"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
}

type TimesheetEntry struct {
	ID            uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	TenantID      uint      `gorm:"not null;index" json:"tenant_id"`
	UserID        uint      `gorm:"not null;index" json:"user_id"`
	ProjectID     uint      `gorm:"not null;index" json:"project_id" binding:"required"`
	TaskID        *uint     `gorm:"index" json:"task_id"`
	Date          time.Time `gorm:"not null;type:date" json:"date" binding:"required"`
	DurationHours float64   `gorm:"not null;type:decimal(5,2)" json:"duration_hours" binding:"required,gt=0"`
	Notes         string    `gorm:"type:text" json:"notes"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	User    *User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Project *Project `gorm:"foreignKey:ProjectID" json:"project,omitempty"`
	Task    *Task    `gorm:"foreignKey:TaskID" json:"task,omitempty"`
}

// DTO for Monthly Report
type MonthlyTimesheetReport struct {
	EmployeeName     string                 `json:"employee_name"`
	EmployeeID       string                 `json:"employee_id"`
	Department       string                 `json:"department"`
	Period           string                 `json:"period"`
	Entries          []TimesheetEntry       `json:"entries"`
	TotalHours       float64                `json:"total_hours"`
	ProjectBreakdown map[string]float64     `json:"project_breakdown"`
	Signatures       TimesheetSignatures    `json:"signatures"`
}

type TimesheetSignatures struct {
	EmployeeName string `json:"employee_name"`
	ManagerName  string `json:"manager_name"`
	HRName       string `json:"hr_name"`
}
