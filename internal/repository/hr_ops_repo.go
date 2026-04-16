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
	FindDefaultShift(ctx context.Context, tenantID uint) (*model.WorkShift, error)
	CreateShift(ctx context.Context, shift *model.WorkShift) error

	// Roster
	FindRoster(ctx context.Context, tenantID uint, userID uint, startDate, endDate time.Time) ([]model.EmployeeRoster, error)
	SaveRoster(ctx context.Context, rosters []model.EmployeeRoster) error

	// Calendar Events
	FindEvents(ctx context.Context, tenantID uint, year int) ([]model.CalendarEvent, error)
	FindEventByDate(ctx context.Context, tenantID uint, date time.Time) (*model.CalendarEvent, error)
	FindHolidayByDate(ctx context.Context, tenantID uint, date time.Time) (*model.CalendarEvent, error) // Legacy for attendance
	FindEventByID(ctx context.Context, tenantID uint, id uuid.UUID) (*model.CalendarEvent, error)
	FindUpcomingEvents(ctx context.Context, date time.Time) ([]model.CalendarEvent, error)
	CreateEvent(ctx context.Context, event *model.CalendarEvent) error
	UpdateEvent(ctx context.Context, event *model.CalendarEvent) error
	DeleteEvent(ctx context.Context, tenantID uint, id uuid.UUID) error

	// Lifecycle
	FindLifecycleTasks(ctx context.Context, tenantID uint, category *model.LifecycleStatus) ([]model.LifecycleTask, error)
	CreateLifecycleTask(ctx context.Context, task *model.LifecycleTask) error
	DeleteLifecycleTask(ctx context.Context, tenantID uint, id uuid.UUID) error
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

func (r *hrOpsRepository) FindDefaultShift(ctx context.Context, tenantID uint) (*model.WorkShift, error) {
	var shift model.WorkShift
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND is_default = ?", tenantID, true).First(&shift).Error
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
				Assign(map[string]interface{}{
					"shift_id":  roster.ShiftID,
					"tenant_id": roster.TenantID,
				}).
				FirstOrCreate(&roster).Error
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *hrOpsRepository) FindEvents(ctx context.Context, tenantID uint, year int) ([]model.CalendarEvent, error) {
	var events []model.CalendarEvent
	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)
	if year != 0 {
		start := time.Date(year, 1, 1, 0, 0, 0, 0, time.UTC)
		end := time.Date(year, 12, 31, 23, 59, 59, 0, time.UTC)
		query = query.Where("date BETWEEN ? AND ?", start, end)
	}
	err := query.Preload("Users").Find(&events).Error
	return events, err
}

func (r *hrOpsRepository) FindEventByDate(ctx context.Context, tenantID uint, date time.Time) (*model.CalendarEvent, error) {
	var event model.CalendarEvent
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND date = ?", tenantID, date.Format("2006-01-02")).Preload("Users").First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *hrOpsRepository) FindHolidayByDate(ctx context.Context, tenantID uint, date time.Time) (*model.CalendarEvent, error) {
	var event model.CalendarEvent
	// Only return if it blocks office (Office Closed)
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND date = ? AND category = ?", 
		tenantID, date.Format("2006-01-02"), model.EventCategoryOfficeClosed).First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *hrOpsRepository) FindEventByID(ctx context.Context, tenantID uint, id uuid.UUID) (*model.CalendarEvent, error) {
	var event model.CalendarEvent
	err := r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Preload("Users").First(&event).Error
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *hrOpsRepository) FindUpcomingEvents(ctx context.Context, date time.Time) ([]model.CalendarEvent, error) {
	var events []model.CalendarEvent
	err := r.db.WithContext(ctx).Where("date = ?", date.Format("2006-01-02")).Preload("Users").Find(&events).Error
	return events, err
}

func (r *hrOpsRepository) CreateEvent(ctx context.Context, event *model.CalendarEvent) error {
	return r.db.WithContext(ctx).Create(event).Error
}

func (r *hrOpsRepository) UpdateEvent(ctx context.Context, event *model.CalendarEvent) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Clear existing associations
		if err := tx.Model(event).Association("Users").Clear(); err != nil {
			return err
		}
		// Save event and associations
		return tx.Save(event).Error
	})
}

func (r *hrOpsRepository) DeleteEvent(ctx context.Context, tenantID uint, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&model.CalendarEvent{}).Error
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

func (r *hrOpsRepository) CreateLifecycleTask(ctx context.Context, task *model.LifecycleTask) error {
	return r.db.WithContext(ctx).Create(task).Error
}

func (r *hrOpsRepository) DeleteLifecycleTask(ctx context.Context, tenantID uint, id uuid.UUID) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND id = ?", tenantID, id).Delete(&model.LifecycleTask{}).Error
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
