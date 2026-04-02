package repository

import (
	"context"
	"errors"
	"time"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type AttendanceRepository interface {
	Save(ctx context.Context, attendance *model.Attendance) error
	Update(ctx context.Context, attendance *model.Attendance) error
	FindTodayByUser(ctx context.Context, userID uint) (*model.Attendance, error)
	FindAll(ctx context.Context, filter model.AttendanceFilter, limit, offset int) ([]model.Attendance, int64, error)
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

func (r *attendanceRepository) FindTodayByUser(ctx context.Context, userID uint) (*model.Attendance, error) {
	var attendance model.Attendance

	startOfDay := time.Now().Truncate(24 * time.Hour)
	endOfDay := startOfDay.Add(24 * time.Hour)

	err := r.db.WithContext(ctx).
		Where("user_id = ? AND clock_in_time >= ? AND clock_in_time < ?", userID, startOfDay, endOfDay).
		First(&attendance).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &attendance, nil
}

func (r *attendanceRepository) FindAll(
	ctx context.Context,
	filter model.AttendanceFilter,
	limit, offset int,
) ([]model.Attendance, int64, error) {

	var attendances []model.Attendance
	var total int64

	query := r.db.WithContext(ctx).
		Model(&model.Attendance{}).
		Preload("User")

	if filter.UserID != 0 {
		query = query.Where("user_id = ?", filter.UserID)
	}

	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}

	if filter.DateFrom != nil {
		query = query.Where("clock_in_time >= ?", *filter.DateFrom)
	}

	if filter.DateTo != nil {
		query = query.Where("clock_in_time <= ?", *filter.DateTo)
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
