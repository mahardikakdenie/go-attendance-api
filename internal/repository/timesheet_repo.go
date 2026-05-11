package repository

import (
	"context"
	"go-attendance-api/internal/model"
	"time"

	"gorm.io/gorm"
)

type TimesheetRepository interface {
	// Project
	CreateProject(ctx context.Context, project *model.Project) error
	FindProjects(ctx context.Context, tenantID uint, status string, search string) ([]model.Project, error)
	FindProjectByID(ctx context.Context, id uint, tenantID uint) (*model.Project, error)
	UpdateProject(ctx context.Context, project *model.Project) error
	DeleteProject(ctx context.Context, id uint, tenantID uint) error

	// Project Members
	AddProjectMember(ctx context.Context, member *model.ProjectMember) error
	RemoveProjectMember(ctx context.Context, projectID uint, userID uint) error
	FindProjectMembers(ctx context.Context, projectID uint) ([]model.ProjectMember, error)

	// Task
	CreateTask(ctx context.Context, task *model.Task) error
	FindTasksByProject(ctx context.Context, projectID uint) ([]model.Task, error)
	FindTaskByID(ctx context.Context, id uint) (*model.Task, error)
	FindTaskByName(ctx context.Context, projectID uint, name string) (*model.Task, error)
	UpdateTask(ctx context.Context, task *model.Task) error

	// Timesheet Entry
	CreateEntry(ctx context.Context, entry *model.TimesheetEntry) error
	FindEntriesByUserPeriod(ctx context.Context, userID uint, month, year int) ([]model.TimesheetEntry, error)
	FindEntriesByUserDateRange(ctx context.Context, userID uint, start, end time.Time) ([]model.TimesheetEntry, error)
	FindEntriesByTenantPeriod(ctx context.Context, tenantID uint, month, year int) ([]model.TimesheetEntry, error)
	FindEntriesByTenantDateRange(ctx context.Context, tenantID uint, start, end time.Time) ([]model.TimesheetEntry, error)
	FindEntries(ctx context.Context, filter model.TimesheetFilter, limit, offset int) ([]model.TimesheetEntry, int64, error)

	// High-Scale Analytics (Aggregated)
	GetAnalyticsSummary(ctx context.Context, filter model.TimesheetFilter) (totalHours float64, activeEmployees int64, err error)
	GetProjectDistribution(ctx context.Context, filter model.TimesheetFilter) ([]struct {
		ProjectID   uint
		ProjectName string
		TotalHours  float64
	}, error)
	GetDailyStats(ctx context.Context, filter model.TimesheetFilter) ([]struct {
		Date       time.Time
		TotalHours float64
	}, error)
}

type timesheetRepository struct {
	db *gorm.DB
}

func NewTimesheetRepository(db *gorm.DB) TimesheetRepository {
	return &timesheetRepository{db: db}
}

func (r *timesheetRepository) CreateProject(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Create(project).Error
}

func (r *timesheetRepository) FindProjects(ctx context.Context, tenantID uint, status string, search string) ([]model.Project, error) {
	var projects []model.Project
	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if search != "" {
		query = query.Where("(name ILIKE ? OR client_name ILIKE ?)", "%"+search+"%", "%"+search+"%")
	}

	err := query.Order("created_at DESC").Find(&projects).Error
	return projects, err
}

func (r *timesheetRepository) FindProjectByID(ctx context.Context, id uint, tenantID uint) (*model.Project, error) {
	var project model.Project
	err := r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).First(&project).Error
	if err != nil {
		return nil, err
	}
	return &project, nil
}

func (r *timesheetRepository) UpdateProject(ctx context.Context, project *model.Project) error {
	return r.db.WithContext(ctx).Save(project).Error
}

func (r *timesheetRepository) DeleteProject(ctx context.Context, id uint, tenantID uint) error {
	return r.db.WithContext(ctx).Where("id = ? AND tenant_id = ?", id, tenantID).Delete(&model.Project{}).Error
}

func (r *timesheetRepository) AddProjectMember(ctx context.Context, member *model.ProjectMember) error {
	return r.db.WithContext(ctx).Save(member).Error
}

func (r *timesheetRepository) RemoveProjectMember(ctx context.Context, projectID uint, userID uint) error {
	return r.db.WithContext(ctx).Where("project_id = ? AND user_id = ?", projectID, userID).Delete(&model.ProjectMember{}).Error
}

func (r *timesheetRepository) FindProjectMembers(ctx context.Context, projectID uint) ([]model.ProjectMember, error) {
	var members []model.ProjectMember
	err := r.db.WithContext(ctx).Preload("User").Where("project_id = ?", projectID).Find(&members).Error
	return members, err
}

func (r *timesheetRepository) CreateTask(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *timesheetRepository) FindTasksByProject(ctx context.Context, projectID uint) ([]model.Task, error) {
	var tasks []model.Task
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&tasks).Error
	return tasks, err
}

func (r *timesheetRepository) FindTaskByID(ctx context.Context, id uint) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).First(&task, id).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *timesheetRepository) FindTaskByName(ctx context.Context, projectID uint, name string) (*model.Task, error) {
	var task model.Task
	err := r.db.WithContext(ctx).Where("project_id = ? AND name = ?", projectID, name).First(&task).Error
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (r *timesheetRepository) UpdateTask(ctx context.Context, task *model.Task) error {
	return r.db.WithContext(ctx).Save(task).Error
}

func (r *timesheetRepository) CreateEntry(ctx context.Context, entry *model.TimesheetEntry) error {
	return r.db.WithContext(ctx).Create(entry).Error
}

func (r *timesheetRepository) FindEntriesByUserPeriod(ctx context.Context, userID uint, month, year int) ([]model.TimesheetEntry, error) {
	var entries []model.TimesheetEntry
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Task").
		Where("user_id = ? AND EXTRACT(MONTH FROM date) = ? AND EXTRACT(YEAR FROM date) = ?", userID, month, year).
		Order("date ASC").
		Find(&entries).Error
	return entries, err
}

func (r *timesheetRepository) FindEntriesByUserDateRange(ctx context.Context, userID uint, start, end time.Time) ([]model.TimesheetEntry, error) {
	var entries []model.TimesheetEntry
	err := r.db.WithContext(ctx).
		Preload("Project").
		Preload("Task").
		Where("user_id = ? AND date BETWEEN ? AND ?", userID, start, end).
		Order("date ASC").
		Find(&entries).Error
	return entries, err
}

func (r *timesheetRepository) FindEntriesByTenantPeriod(ctx context.Context, tenantID uint, month, year int) ([]model.TimesheetEntry, error) {
	var entries []model.TimesheetEntry
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Project").
		Preload("Task").
		Where("tenant_id = ? AND EXTRACT(MONTH FROM date) = ? AND EXTRACT(YEAR FROM date) = ?", tenantID, month, year).
		Order("user_id, date ASC").
		Find(&entries).Error
	return entries, err
}

func (r *timesheetRepository) FindEntriesByTenantDateRange(ctx context.Context, tenantID uint, start, end time.Time) ([]model.TimesheetEntry, error) {
	var entries []model.TimesheetEntry
	err := r.db.WithContext(ctx).
		Preload("User").
		Preload("Project").
		Preload("Task").
		Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, start, end).
		Order("user_id, date ASC").
		Find(&entries).Error
	return entries, err
}

func (r *timesheetRepository) FindEntries(ctx context.Context, filter model.TimesheetFilter, limit, offset int) ([]model.TimesheetEntry, int64, error) {
	var entries []model.TimesheetEntry
	var total int64

	query := r.db.WithContext(ctx).Model(&model.TimesheetEntry{}).Where("tenant_id = ?", filter.TenantID)

	if filter.UserID != 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if filter.ProjectID != 0 {
		query = query.Where("project_id = ?", filter.ProjectID)
	}

	if filter.StartDate != nil && filter.EndDate != nil {
		query = query.Where("date BETWEEN ? AND ?", filter.StartDate, filter.EndDate)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch data
	err := query.
		Preload("User").
		Preload("Project").
		Preload("Task").
		Limit(limit).
		Offset(offset).
		Order("date DESC, created_at DESC").
		Find(&entries).Error

	return entries, total, err
}

func (r *timesheetRepository) GetAnalyticsSummary(ctx context.Context, filter model.TimesheetFilter) (totalHours float64, activeEmployees int64, err error) {
	query := r.db.WithContext(ctx).Model(&model.TimesheetEntry{}).Where("tenant_id = ?", filter.TenantID)

	if filter.UserID != 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if filter.ProjectID != 0 {
		query = query.Where("project_id = ?", filter.ProjectID)
	}

	if filter.StartDate != nil && filter.EndDate != nil {
		query = query.Where("date BETWEEN ? AND ?", filter.StartDate, filter.EndDate)
	}

	// Calculate total hours
	if err := query.Select("COALESCE(SUM(duration_hours), 0)").Scan(&totalHours).Error; err != nil {
		return 0, 0, err
	}

	// Calculate active employees
	if err := query.Select("COUNT(DISTINCT user_id)").Scan(&activeEmployees).Error; err != nil {
		return 0, 0, err
	}

	return totalHours, activeEmployees, nil
}

func (r *timesheetRepository) GetProjectDistribution(ctx context.Context, filter model.TimesheetFilter) ([]struct {
	ProjectID   uint
	ProjectName string
	TotalHours  float64
}, error) {
	var results []struct {
		ProjectID   uint
		ProjectName string
		TotalHours  float64
	}

	query := r.db.WithContext(ctx).
		Table("timesheet_entries").
		Select("timesheet_entries.project_id, projects.name as project_name, SUM(timesheet_entries.duration_hours) as total_hours").
		Joins("JOIN projects ON projects.id = timesheet_entries.project_id").
		Where("timesheet_entries.tenant_id = ?", filter.TenantID)

	if filter.UserID != 0 {
		query = query.Where("timesheet_entries.user_id = ?", filter.UserID)
	}

	if filter.StartDate != nil && filter.EndDate != nil {
		query = query.Where("timesheet_entries.date BETWEEN ? AND ?", filter.StartDate, filter.EndDate)
	}

	err := query.
		Group("timesheet_entries.project_id, projects.name").
		Order("total_hours DESC").
		Scan(&results).Error

	return results, err
}

func (r *timesheetRepository) GetDailyStats(ctx context.Context, filter model.TimesheetFilter) ([]struct {
	Date       time.Time
	TotalHours float64
}, error) {
	var results []struct {
		Date       time.Time
		TotalHours float64
	}

	query := r.db.WithContext(ctx).
		Table("timesheet_entries").
		Select("date, SUM(duration_hours) as total_hours").
		Where("tenant_id = ?", filter.TenantID)

	if filter.UserID != 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if filter.StartDate != nil && filter.EndDate != nil {
		query = query.Where("date BETWEEN ? AND ?", filter.StartDate, filter.EndDate)
	}

	err := query.
		Group("date").
		Order("date ASC").
		Scan(&results).Error

	return results, err
}
