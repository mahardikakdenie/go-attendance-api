package repository

import (
	"context"
	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type OvertimeRepository interface {
	Save(ctx context.Context, overtime *model.Overtime) error
	Update(ctx context.Context, overtime *model.Overtime) error
	FindByID(ctx context.Context, id uint) (*model.Overtime, error)
	FindAll(ctx context.Context, filter model.OvertimeFilter) ([]model.Overtime, int64, error)
	GetPendingCount(ctx context.Context, userID uint) (int64, error)
}

type overtimeRepository struct {
	db *gorm.DB
}

func NewOvertimeRepository(db *gorm.DB) OvertimeRepository {
	return &overtimeRepository{
		db: db,
	}
}

func (r *overtimeRepository) GetPendingCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Overtime{}).
		Where("overtimes.user_id = ? AND overtimes.status = ?", userID, model.OvertimeStatusPending).
		Count(&count).Error
	return count, err
}

func (r *overtimeRepository) Save(ctx context.Context, overtime *model.Overtime) error {
	return r.db.WithContext(ctx).Create(overtime).Error
}

func (r *overtimeRepository) Update(ctx context.Context, overtime *model.Overtime) error {
	return r.db.WithContext(ctx).Save(overtime).Error
}

func (r *overtimeRepository) FindByID(ctx context.Context, id uint) (*model.Overtime, error) {
	var overtime model.Overtime
	err := r.db.WithContext(ctx).Preload("User").First(&overtime, id).Error
	if err != nil {
		return nil, err
	}
	return &overtime, nil
}

func (r *overtimeRepository) FindAll(ctx context.Context, filter model.OvertimeFilter) ([]model.Overtime, int64, error) {
	var overtimes []model.Overtime
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Overtime{})

	if filter.TenantID != 0 {
		query = query.Where("overtimes.tenant_id = ?", filter.TenantID)
	}

	if filter.UserID != 0 {
		query = query.Where("overtimes.user_id = ?", filter.UserID)
	}

	if filter.Status != "" {
		query = query.Where("overtimes.status = ?", filter.Status)
	}

	if filter.DateFrom != nil {
		query = query.Where("date >= ?", *filter.DateFrom)
	}

	if filter.DateTo != nil {
		query = query.Where("date <= ?", *filter.DateTo)
	}

	if len(filter.AllowedRoleIDs) > 0 {
		query = query.Joins("User").Where("\"User\".role_id IN ?", filter.AllowedRoleIDs)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit).Offset(filter.Offset)
	}

	err := query.Order("created_at DESC").Preload("User").Find(&overtimes).Error
	if err != nil {
		return nil, 0, err
	}

	return overtimes, total, nil
}
