package repository

import (
	"context"
	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type SubscriptionRepository interface {
	FindAll(ctx context.Context, page, limit int, status, search string) ([]model.Subscription, int64, error)
	FindByID(ctx context.Context, id uint) (*model.Subscription, error)
	Create(ctx context.Context, sub *model.Subscription) error
	Update(ctx context.Context, sub *model.Subscription) error
	GetStats(ctx context.Context) (float64, int64, float64, error)
	CountEmployees(ctx context.Context, tenantID uint) (int64, error)
	FindByTenantID(ctx context.Context, tenantID uint) (*model.Subscription, error)
	FindPlanByName(ctx context.Context, name string) (*model.SubscriptionPlan, error)

	// SubscriptionPlan CRUD
	FindAllPlans(ctx context.Context) ([]model.SubscriptionPlan, error)
	FindPlanByID(ctx context.Context, id uint) (*model.SubscriptionPlan, error)
	CreatePlan(ctx context.Context, plan *model.SubscriptionPlan) error
	UpdatePlan(ctx context.Context, plan *model.SubscriptionPlan) error
	DeletePlan(ctx context.Context, id uint) error
}

type subscriptionRepository struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) SubscriptionRepository {
	return &subscriptionRepository{db: db}
}

func (r *subscriptionRepository) Create(ctx context.Context, sub *model.Subscription) error {
	return r.db.WithContext(ctx).Create(sub).Error
}

func (r *subscriptionRepository) FindAll(ctx context.Context, page, limit int, status, search string) ([]model.Subscription, int64, error) {
	var subs []model.Subscription
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Subscription{}).Preload("Tenant").Preload("Tenant.TenantSettings")

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if search != "" {
		query = query.Joins("JOIN tenants ON subscriptions.tenant_id = tenants.id").
			Where("tenants.name ILIKE ? OR tenants.code ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit
	err = query.Offset(offset).Limit(limit).Find(&subs).Error

	return subs, total, err
}

func (r *subscriptionRepository) FindByID(ctx context.Context, id uint) (*model.Subscription, error) {
	var sub model.Subscription
	err := r.db.WithContext(ctx).Preload("Tenant").First(&sub, id).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) Update(ctx context.Context, sub *model.Subscription) error {
	return r.db.WithContext(ctx).Save(sub).Error
}

func (r *subscriptionRepository) GetStats(ctx context.Context) (float64, int64, float64, error) {
	var mrr float64
	var activeTenants int64
	var pastDueAmount float64

	// MRR: Sum of Active/Trial
	r.db.WithContext(ctx).Model(&model.Subscription{}).
		Where("status IN ?", []model.SubscriptionStatus{model.SubscriptionStatusActive, model.SubscriptionStatusTrial}).
		Select("COALESCE(SUM(amount), 0)").Scan(&mrr)

	// Active Tenants
	r.db.WithContext(ctx).Model(&model.Subscription{}).
		Where("status = ?", model.SubscriptionStatusActive).
		Count(&activeTenants)

	// Past Due Amount
	r.db.WithContext(ctx).Model(&model.Subscription{}).
		Where("status = ?", model.SubscriptionStatusPastDue).
		Select("COALESCE(SUM(amount), 0)").Scan(&pastDueAmount)

	return mrr, activeTenants, pastDueAmount, nil
}

func (r *subscriptionRepository) CountEmployees(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	r.db.WithContext(ctx).Model(&model.User{}).Where("tenant_id = ?", tenantID).Count(&count)
	return count, nil
}

func (r *subscriptionRepository) FindByTenantID(ctx context.Context, tenantID uint) (*model.Subscription, error) {
	var sub model.Subscription
	err := r.db.WithContext(ctx).Preload("Tenant").Where("tenant_id = ?", tenantID).First(&sub).Error
	if err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *subscriptionRepository) FindPlanByName(ctx context.Context, name string) (*model.SubscriptionPlan, error) {
	var plan model.SubscriptionPlan
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *subscriptionRepository) FindAllPlans(ctx context.Context) ([]model.SubscriptionPlan, error) {
	var plans []model.SubscriptionPlan
	err := r.db.WithContext(ctx).Order("id ASC").Find(&plans).Error
	return plans, err
}

func (r *subscriptionRepository) FindPlanByID(ctx context.Context, id uint) (*model.SubscriptionPlan, error) {
	var plan model.SubscriptionPlan
	err := r.db.WithContext(ctx).First(&plan, id).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *subscriptionRepository) CreatePlan(ctx context.Context, plan *model.SubscriptionPlan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *subscriptionRepository) UpdatePlan(ctx context.Context, plan *model.SubscriptionPlan) error {
	return r.db.WithContext(ctx).Save(plan).Error
}

func (r *subscriptionRepository) DeletePlan(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.SubscriptionPlan{}, id).Error
}
