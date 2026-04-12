package repository

import (
	"context"
	"go-attendance-api/internal/model"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type HrOpsRepository interface {
	// Shifts
	FindAllShifts(ctx context.Context, tenantID uint) ([]model.WorkShift, error)
	FindShiftByID(ctx context.Context, tenantID uint, id uuid.UUID) (*model.WorkShift, error)
	CreateShift(ctx context.Context, shift *model.WorkShift) error

	// Roster
	FindRoster(ctx context.Context, tenantID uint, userID uint, startDate, endDate time.Time) ([]model.EmployeeRoster, error)
	SaveRoster(ctx context.Context, rosters []model.EmployeeRoster) error

	// Holidays
	FindHolidays(ctx context.Context, tenantID uint, year int) ([]model.Holiday, error)
	FindHolidayByDate(ctx context.Context, tenantID uint, date time.Time) (*model.Holiday, error)
	FindHolidayByID(ctx context.Context, tenantID uint, id uuid.UUID) (*model.Holiday, error)
	CreateHoliday(ctx context.Context, holiday *model.Holiday) error
	UpdateHoliday(ctx context.Context, holiday *model.Holiday) error
	DeleteHoliday(ctx context.Context, tenantID uint, id uuid.UUID) error

	// Lifecycle
	FindLifecycleTasks(ctx context.Context, tenantID uint, category *model.LifecycleStatus) ([]model.LifecycleTask, error)
	FindEmployeeLifecycle(ctx context.Context, userID uint) ([]model.EmployeeLifecycleTask, error)
	UpdateEmployeeLifecycleTask(ctx context.Context, userID uint, taskID uuid.UUID, isCompleted bool) error
}

type hrOpsRepository struct {
	db *gorm.DB
}

func NewHrOpsRepository(db *gorm.DB) HrOpsRepository {
	return &hrOpsRepository{db: db}
}

func (r *hrOpsRepository) FindAllShifts(ctx context.Context, tenantID uint) ([]model.WorkShift, error) {
	var shifts []model.WorkShift
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&shifts).Error
	return shifts, err
}

func (r *hrOpsRepository) FindShiftByID(ctx context.Context, tenantID uint, id uuid.UUID) (*model.WorkShift, error) {
	var shift model.WorkShift
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&shift).Error
	if err != nil {
		return nil, err
	}
	return &shift, nil
}

func (r *hrOpsRepository) CreateShift(ctx context.Context, shift *model.WorkShift) error {
	return r.db.WithContext(ctx).Create(shift).Error
}

func (r *hrOpsRepository) FindRoster(ctx context.Context, tenantID uint, userID uint, startDate, endDate time.Time) ([]model.EmployeeRoster, error) {
	var rosters []model.EmployeeRoster
	query := r.db.WithContext(ctx).Where("tenant_id = ? AND date BETWEEN ? AND ?", tenantID, startDate, endDate)
	if userID != 0 {
		query = query.Where("user_id = ?", userID)
	}
	err := query.Preload("Shift").Find(&rosters).Error
	return rosters, err
}

func (r *hrOpsRepository) SaveRoster(ctx context.Context, rosters []model.EmployeeRoster) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, roster := range rosters {
			err := tx.Where("user_id = ? AND date = ?", roster.UserID, roster.Date).
				Assign(model.EmployeeRoster{ShiftID: roster.ShiftID, TenantID: roster.TenantID}).
				FirstOrCreate(&roster).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *hrOpsRepository) FindHolidays(ctx context.Context, tenantID uint, year int) ([]model.Holiday, error) {
	var holidays []model.Holiday
	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if year != 0 {
		start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)
		query = query.Where("date BETWEEN ? AND ?", start, end)
	}
	err := query.Find(&holidays).Error
	return holidays, err
}

func (r *hrOpsRepository) FindHolidayByDate(ctx context.Context, tenantID uint, date time.Time) (*model.Holiday, error) {
	var holiday model.Holiday
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND date = ?", tenantID, date.Format("2006-01-02")).First(&holiday).Error
	if err != nil {
		return nil, err
	}
	return &holiday, nil
}

func (r *hrOpsRepository) FindHolidayByID(ctx context.Context, tenantID uint, id uuid.UUID) (*model.Holiday, error) {
	var holiday model.Holiday
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).First(&holiday).Error
	if err != nil {
		return nil, err
	}
	return &holiday, nil
}

func (r *hrOpsRepository) CreateHoliday(ctx context.Context, holiday *model.Holiday) error {
	return r.db.WithContext(ctx).Create(holiday).Error
}

func (r *hrOpsRepository) UpdateHoliday(ctx context.Context, holiday *model.Holiday) error {
	return r.db.WithContext(ctx).Save(holiday).Error
}

func (r *hrOpsRepository) DeleteHoliday(ctx context.Context, tenantID uint, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&model.Holiday{}).Error
}

func (r *hrOpsRepository) FindLifecycleTasks(ctx context.Context, tenantID uint, category *model.LifecycleStatus) ([]model.LifecycleTask, error) {
	var tasks []model.LifecycleTask
	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if category != nil {
		query = query.Where("category = ?", *category)
	}
	err := query.Find(&tasks).Error
	return tasks, err
}

func (r *hrOpsRepository) FindEmployeeLifecycle(ctx context.Context, userID uint) ([]model.EmployeeLifecycleTask, error) {
	var tasks []model.EmployeeLifecycleTask
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Preload("Task").Find(&tasks).Error
	return tasks, err
}

func (r *hrOpsRepository) UpdateEmployeeLifecycleTask(ctx context.Context, userID uint, taskID uuid.UUID, isCompleted bool) error {
	now := time.Now()
	var completedAt *time.Time
	if isCompleted {
		completedAt = &now
	}

	return r.db.WithContext(ctx).Model(&model.EmployeeLifecycleTask{}).
		Where("user_id = ? AND task_id = ?", userID, taskID).
		Updates(map[string]interface{}{
			"is_completed": isCompleted,
			"completed_at": completedAt,
		}).Error
}
