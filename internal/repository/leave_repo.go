package repository

import (
	"context"
	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type LeaveRepository interface {
	CreateLeaveType(ctx context.Context, lt *model.LeaveType) error
	GetLeaveTypesByTenant(ctx context.Context, tenantID uint) ([]model.LeaveType, error)
	GetLeaveTypeByID(ctx context.Context, id uint) (*model.LeaveType, error)

	GetBalance(ctx context.Context, userID uint, leaveTypeID uint, year int) (*model.LeaveBalance, error)
	UpdateBalance(ctx context.Context, lb *model.LeaveBalance) error
	CreateBalance(ctx context.Context, lb *model.LeaveBalance) error

	CreateLeave(ctx context.Context, l *model.Leave) error
	GetLeavesByUser(ctx context.Context, userID uint, limit, offset int) ([]model.Leave, int64, error)
	GetPendingCount(ctx context.Context, userID uint) (int64, error)
}

type leaveRepository struct {
	db *gorm.DB
}

func NewLeaveRepository(db *gorm.DB) LeaveRepository {
	return &leaveRepository{db: db}
}

func (r *leaveRepository) GetPendingCount(ctx context.Context, userID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Leave{}).
		Where("user_id = ? AND status = ?", userID, model.LeaveStatusPending).
		Count(&count).Error
	return count, err
}

func (r *leaveRepository) CreateLeaveType(ctx context.Context, lt *model.LeaveType) error {
	return r.db.WithContext(ctx).Create(lt).Error
}

func (r *leaveRepository) GetLeaveTypesByTenant(ctx context.Context, tenantID uint) ([]model.LeaveType, error) {
	var results []model.LeaveType
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Find(&results).Error
	return results, err
}

func (r *leaveRepository) GetLeaveTypeByID(ctx context.Context, id uint) (*model.LeaveType, error) {
	var result model.LeaveType
	err := r.db.WithContext(ctx).First(&result, id).Error
	return &result, err
}

func (r *leaveRepository) GetBalance(ctx context.Context, userID uint, leaveTypeID uint, year int) (*model.LeaveBalance, error) {
	var result model.LeaveBalance
	err := r.db.WithContext(ctx).Where("user_id = ? AND leave_type_id = ? AND year = ?", userID, leaveTypeID, year).First(&result).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return &result, err
}

func (r *leaveRepository) UpdateBalance(ctx context.Context, lb *model.LeaveBalance) error {
	return r.db.WithContext(ctx).Save(lb).Error
}

func (r *leaveRepository) CreateBalance(ctx context.Context, lb *model.LeaveBalance) error {
	return r.db.WithContext(ctx).Create(lb).Error
}

func (r *leaveRepository) CreateLeave(ctx context.Context, l *model.Leave) error {
	return r.db.WithContext(ctx).Create(l).Error
}

func (r *leaveRepository) GetLeavesByUser(ctx context.Context, userID uint, limit, offset int) ([]model.Leave, int64, error) {
	var results []model.Leave
	var total int64
	query := r.db.WithContext(ctx).Model(&model.Leave{}).Where("user_id = ?", userID)
	query.Count(&total)
	err := query.Order("created_at DESC").Limit(limit).Offset(offset).Preload("LeaveType").Find(&results).Error
	return results, total, err
}
