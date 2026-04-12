package modelDto

import (
	"time"

	"go-attendance-api/internal/model"

	"github.com/google/uuid"
)

// Shift DTOs
type WorkShiftResponse struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	StartTime string          `json:"startTime"`
	EndTime   string          `json:"endTime"`
	Type      model.ShiftType `json:"type"`
	Color     string          `json:"color"`
	IsDefault bool            `json:"isDefault"`
}

// Roster DTOs
type EmployeeScheduleResponse struct {
	ID           uint              `json:"id"`
	Name         string            `json:"name"`
	Avatar       string            `json:"avatar"`
	Department   string            `json:"department"`
	WeeklyRoster map[string]string `json:"weeklyRoster"` // day -> shift_id or "off"
}

type RosterAssignment struct {
	UserID uint              `json:"user_id" binding:"required"`
	Roster map[string]string `json:"roster" binding:"required"` // day -> shift_id or "off"
}

type SaveRosterRequest struct {
	StartDate   string             `json:"start_date" binding:"required"` // YYYY-MM-DD
	Assignments []RosterAssignment `json:"assignments" binding:"required"`
}

// Holiday DTOs
type HolidayResponse struct {
	ID     uuid.UUID         `json:"id"`
	Date   string            `json:"date"` // YYYY-MM-DD
	Name   string            `json:"name"`
	Type   model.HolidayType `json:"type"`
	IsPaid bool              `json:"is_paid"`
}

type CreateHolidayRequest struct {
	Date   string            `json:"date" binding:"required"` // YYYY-MM-DD
	Name   string            `json:"name" binding:"required"`
	Type   model.HolidayType `json:"type" binding:"required"`
	IsPaid bool              `json:"is_paid"`
}

type UpdateHolidayRequest struct {
	Name   string `json:"name"`
	IsPaid *bool  `json:"is_paid"`
}

// Lifecycle DTOs
type LifecycleTaskResponse struct {
	ID          uuid.UUID             `json:"id"`
	TaskName    string                `json:"task_name"`
	Category    model.LifecycleStatus `json:"category"`
	IsCompleted bool                  `json:"is_completed"`
	CompletedAt *time.Time            `json:"completed_at"`
}

type EmployeeLifecycleResponse struct {
	EmployeeID uint                    `json:"employee_id"`
	Status     model.LifecycleStatus   `json:"status"`
	Tasks      []LifecycleTaskResponse `json:"tasks"`
}

type UpdateLifecycleTaskRequest struct {
	IsCompleted bool `json:"is_completed"`
}
