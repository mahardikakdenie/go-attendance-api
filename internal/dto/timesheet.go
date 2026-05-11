package modelDto

import "time"

type TimesheetMonitoringFilter struct {
	StartDate string `form:"start_date" binding:"required"`
	EndDate   string `form:"end_date" binding:"required"`
	UserID    uint   `form:"user_id"`
	ProjectID uint   `form:"project_id"`
	Page      int    `form:"page,default=1"`
	Limit     int    `form:"limit,default=10"`
}

type TimesheetMonitoringResponse struct {
	ID            string      `json:"id"`
	User          MappedUser  `json:"user"`
	Project       ProjectItem `json:"project"`
	TaskName      string      `json:"task_name"`
	Description   string      `json:"description"`
	DurationHours float64     `json:"duration_hours"`
	Date          time.Time   `json:"date"`
}

type PaginatedTimesheetReport struct {
	Entries    []TimesheetEntryDTO `json:"entries"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	Limit      int                 `json:"limit"`
	TotalHours float64             `json:"total_hours"`
}

type TimesheetEntryDTO struct {
	ID            string    `json:"id"`
	ProjectName   string    `json:"project_name"`
	TaskName      string    `json:"task_name"`
	Description   string    `json:"description"`
	DurationHours float64   `json:"duration_hours"`
	Date          time.Time `json:"date"`
	CreatedAt     time.Time `json:"created_at"`
}

type ProjectItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type TimesheetAnalyticsFilter struct {
	StartDate string `form:"start_date" binding:"required"`
	EndDate   string `form:"end_date" binding:"required"`
	UserID    uint   `form:"user_id"`
}

type TimesheetAnalyticsResponse struct {
	TotalHours          float64                    `json:"total_hours"`
	ActiveEmployees     int64                      `json:"active_employees"`
	ProjectDistribution []ProjectDistributionStats `json:"project_distribution"`
	DailyStats          []DailyTimesheetStats      `json:"daily_stats"`
}

type ProjectDistributionStats struct {
	ProjectID   uint    `json:"project_id"`
	ProjectName string  `json:"project_name"`
	TotalHours  float64 `json:"total_hours"`
	Percentage  float64 `json:"percentage"`
}

type DailyTimesheetStats struct {
	Date       string  `json:"date"`
	TotalHours float64 `json:"total_hours"`
}
