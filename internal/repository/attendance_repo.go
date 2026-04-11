package repository

import (
	"context"
	"errors"
	"time"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/utils"

	"gorm.io/gorm"
)

type AttendanceRepository interface {
	Save(ctx context.Context, attendance *model.Attendance) error
	Update(ctx context.Context, attendance *model.Attendance) error
	FindTodayByUser(ctx context.Context, userID uint, today time.Time) (*model.Attendance, error)
	FindAll(ctx context.Context, filter model.AttendanceFilter, includes []string, limit, offset int) ([]model.Attendance, int64, error)
	GetSummaryCounts(ctx context.Context, tenantID uint, startTime, endTime time.Time) (map[model.AttendanceStatus]int64, error)
	GetOldestDataDate(ctx context.Context, tenantID uint) (*time.Time, error)
}

type attendanceRepository struct {
	db *gorm.DB
}

func NewAttendanceRepository(db *gorm.DB) AttendanceRepository {
	return &attendanceRepository{
		db: db,
	}
}

func (r *attendanceRepository) Save(ctx context.Context, attendance *model.Attendance) error {
	return r.db.WithContext(ctx).Create(attendance).Error
}

func (r *attendanceRepository) Update(ctx context.Context, attendance *model.Attendance) error {
	return r.db.WithContext(ctx).Save(attendance).Error
}

func (r *attendanceRepository) GetSummaryCounts(ctx context.Context, tenantID uint, startTime, endTime time.Time) (map[model.AttendanceStatus]int64, error) {
	var results []struct {
		Status model.AttendanceStatus
		Count  int64
	}

	query := r.db.WithContext(ctx).Model(&model.Attendance{}).
		Select("status, count(*) as count").
		Where("clock_in_time >= ? AND clock_in_time < ?", startTime, endTime)

	if tenantID != 0 {
		query = query.Where("attendances.tenant_id = ?", tenantID)
	}

	err := query.Group("status").Scan(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[model.AttendanceStatus]int64)
	for _, res := range results {
		counts[res.Status] = res.Count
	}

	return counts, nil
}

func (r *attendanceRepository) GetOldestDataDate(ctx context.Context, tenantID uint) (*time.Time, error) {
	var attendance model.Attendance
	query := r.db.WithContext(ctx).Model(&model.Attendance{}).Order("clock_in_time ASC")

	if tenantID != 0 {
		query = query.Where("attendances.tenant_id = ?", tenantID)
	}

	err := query.First(&attendance).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &attendance.ClockInTime, nil
}

func (r *attendanceRepository) FindTodayByUser(ctx context.Context, userID uint, today time.Time) (*model.Attendance, error) {
	var attendance model.Attendance

	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND clock_in_time >= ? AND clock_in_time < ?", userID, startOfDay, endOfDay).
		Order("clock_in_time DESC").
		First(&attendance).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &attendance, nil
}

var attendancePreloadMap = map[string]string{
	"user":    "User",
	"tenant":  "Tenant",
	"setting": "TenantSetting",
}

func (r *attendanceRepository) FindAll(
	ctx context.Context,
	filter model.AttendanceFilter,
	includes []string,
	limit, offset int,
) ([]model.Attendance, int64, error) {

	var attendances []model.Attendance
	var total int64

	query := r.db.WithContext(ctx).
		Model(&model.Attendance{})

	query = utils.ApplyPreloads(query, includes, attendancePreloadMap)

	if filter.UserID != 0 {
		query = query.Where("attendances.user_id = ?", filter.UserID)
	}

	if filter.TenantID != 0 {
		query = query.Where("attendances.tenant_id = ?", filter.TenantID)
	}

	if filter.Status != "" {
		query = query.Where("attendances.status = ?", filter.Status)
	}

	if filter.DateFrom != nil {
		query = query.Where("clock_in_time >= ?", *filter.DateFrom)
	}

	if filter.DateTo != nil {
		query = query.Where("clock_in_time <= ?", *filter.DateTo)
	}

	if len(filter.AllowedRoleIDs) > 0 {
		query = query.Joins("User").Where("\"User\".role_id IN ?", filter.AllowedRoleIDs)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Order("clock_in_time DESC").Find(&attendances).Error
	if err != nil {
		return nil, 0, err
	}

	return attendances, total, nil
}
