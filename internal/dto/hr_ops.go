package modelDto

import (
	"time"

	"go-attendance-api/internal/model"

	"github.com/google/uuid"
)

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

// Calendar Event DTOs
type CalendarEventResponse struct {
	ID          uuid.UUID           `json:"id"`
	Date        string              `json:"date"` // YYYY-MM-DD
	Name        string              `json:"name"`
	Type        model.EventType     `json:"type"`
	Category    model.EventCategory `json:"category"`
	IsPaid      bool                `json:"is_paid"`
	Description string              `json:"description"`
	IsAllUsers  bool                `json:"is_all_users"`
	UserIDs     []uint              `json:"user_ids,omitempty"`
}

// Legacy alias for FE compatibility
type HolidayResponse = CalendarEventResponse

type CreateCalendarEventRequest struct {
	Date        string              `json:"date" binding:"required"` // YYYY-MM-DD
	Name        string              `json:"name" binding:"required"`
	Type        model.EventType     `json:"type" binding:"required"`
	Category    model.EventCategory `json:"category"`
	IsPaid      bool                `json:"is_paid"`
	Description string              `json:"description"`
	IsAllUsers  bool                `json:"is_all_users"`
	UserIDs     []uint              `json:"user_ids"`
}

type CreateHolidayRequest = CreateCalendarEventRequest

type UpdateCalendarEventRequest struct {
	Name        string               `json:"name"`
	Category    *model.EventCategory `json:"category"`
	IsPaid      *bool                `json:"is_paid"`
	Description string               `json:"description"`
	IsAllUsers  *bool                `json:"is_all_users"`
	UserIDs     []uint               `json:"user_ids"`
}

type UpdateHolidayRequest = UpdateCalendarEventRequest

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

type CreateLifecycleTemplateRequest struct {
	TaskName string                `json:"task_name" binding:"required"`
	Category model.LifecycleStatus `json:"category" binding:"required"`
}
