package repository

import (
	"context"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/utils"

	"gorm.io/gorm"
)

type AttendanceCorrectionRepository interface {
	Create(ctx context.Context, correction *model.AttendanceCorrection) error
	FindByID(ctx context.Context, id uint, includes []string) (*model.AttendanceCorrection, error)
	FindAll(ctx context.Context, tenantID uint, userID uint, status string, limit, offset int) ([]model.AttendanceCorrection, int64, error)
	Update(ctx context.Context, correction *model.AttendanceCorrection) error
}

type attendanceCorrectionRepository struct {
	db *gorm.DB
}

func NewAttendanceCorrectionRepository(db *gorm.DB) AttendanceCorrectionRepository {
	return &attendanceCorrectionRepository{db: db}
}

var correctionPreloadMap = map[string]string{
	"user":       "User",
	"attendance": "Attendance",
	"approver":   "Approver",
}

func (r *attendanceCorrectionRepository) Create(ctx context.Context, correction *model.AttendanceCorrection) error {
	return r.db.WithContext(ctx).Create(correction).Error
}

func (r *attendanceCorrectionRepository) FindByID(ctx context.Context, id uint, includes []string) (*model.AttendanceCorrection, error) {
	var correction model.AttendanceCorrection
	query := r.db.WithContext(ctx)
	query = utils.ApplyPreloads(query, includes, correctionPreloadMap)
	err := query.First(&correction, id).Error
	if err != nil {
		return nil, err
	}
	return &correction, nil
}

func (r *attendanceCorrectionRepository) FindAll(ctx context.Context, tenantID uint, userID uint, status string, limit, offset int) ([]model.AttendanceCorrection, int64, error) {
	var corrections []model.AttendanceCorrection
	var total int64
	query := r.db.WithContext(ctx).Model(&model.AttendanceCorrection{}).Where("tenant_id = ?", tenantID)

	if userID != 0 {
		query = query.Where("user_id = ?", userID)
	}

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if limit > 0 {
		query = query.Limit(limit).Offset(offset)
	}

	err := query.Preload("User").Order("created_at DESC").Find(&corrections).Error
	return corrections, total, err
}

func (r *attendanceCorrectionRepository) Update(ctx context.Context, correction *model.AttendanceCorrection) error {
	return r.db.WithContext(ctx).Save(correction).Error
}
