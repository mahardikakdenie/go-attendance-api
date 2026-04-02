package repository

import (
	"context"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type TenantSettingRepository interface {
	Create(ctx context.Context, setting *model.TenantSetting) error
	Update(ctx context.Context, setting *model.TenantSetting) error
	FindByTenantID(ctx context.Context, tenantID uint) (*model.TenantSetting, error)
}

type tenantSettingRepository struct {
	db *gorm.DB
}

func NewTenantSettingRepository(db *gorm.DB) TenantSettingRepository {
	return &tenantSettingRepository{db: db}
}

func (r *tenantSettingRepository) Create(ctx context.Context, setting *model.TenantSetting) error {
	return r.db.WithContext(ctx).Create(setting).Error
}

func (r *tenantSettingRepository) Update(ctx context.Context, setting *model.TenantSetting) error {
	return r.db.WithContext(ctx).Save(setting).Error
}

func (r *tenantSettingRepository) FindByTenantID(ctx context.Context, tenantID uint) (*model.TenantSetting, error) {
	var setting model.TenantSetting
	err := r.db.WithContext(ctx).
		Where("tenant_id = ?", tenantID).
		First(&setting).Error

	if err != nil {
		return nil, err
	}

	return &setting, nil
}
