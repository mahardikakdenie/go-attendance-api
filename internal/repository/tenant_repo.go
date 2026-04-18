package repository

import (
	"context"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type TenantRepository interface {
	Create(ctx context.Context, tenant *model.Tenant) error
	FindAll(ctx context.Context) ([]model.Tenant, error)
	FindByID(ctx context.Context, id uint) (*model.Tenant, error)
	Update(ctx context.Context, tenant *model.Tenant) error
}

type tenantRepository struct {
	db *gorm.DB
}

func NewTenantRepository(db *gorm.DB) TenantRepository {
	return &tenantRepository{db: db}
}

func (r *tenantRepository) Create(ctx context.Context, tenant *model.Tenant) error {
	return r.db.WithContext(ctx).Create(tenant).Error
}

func (r *tenantRepository) FindAll(ctx context.Context) ([]model.Tenant, error) {
	var tenants []model.Tenant
	err := r.db.WithContext(ctx).Find(&tenants).Error
	return tenants, err
}

func (r *tenantRepository) FindByID(ctx context.Context, id uint) (*model.Tenant, error) {
	var tenant model.Tenant
	err := r.db.WithContext(ctx).Preload("TenantSettings").First(&tenant, id).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *tenantRepository) Update(ctx context.Context, tenant *model.Tenant) error {
	return r.db.WithContext(ctx).Save(tenant).Error
}
