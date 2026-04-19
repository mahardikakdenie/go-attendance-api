package repository

import (
	"context"
	"go-attendance-api/internal/model"

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
	UpdateTask(ctx context.Context, task *model.Task) error

	// Timesheet Entry
	CreateEntry(ctx context.Context, entry *model.TimesheetEntry) error
	FindEntriesByUserPeriod(ctx context.Context, userID uint, month, year int) ([]model.TimesheetEntry, error)
	FindEntriesByTenantPeriod(ctx context.Context, tenantID uint, month, year int) ([]model.TimesheetEntry, error)
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
