package repository

import (
	"context"
	"go-attendance-api/internal/model"
	"time"

	"gorm.io/gorm"
)

type ExpenseRepository interface {
	FindAll(ctx context.Context, filter model.ExpenseFilter) ([]model.Expense, int64, error)
	FindByID(ctx context.Context, id uint) (*model.Expense, error)
	Create(ctx context.Context, expense *model.Expense) error
	Update(ctx context.Context, expense *model.Expense) error
	GetSummary(ctx context.Context, tenantID uint) (float64, float64, string, float64, error)
}

type expenseRepository struct {
	db *gorm.DB
}

func NewExpenseRepository(db *gorm.DB) ExpenseRepository {
	return &expenseRepository{db: db}
}

func (r *expenseRepository) FindAll(ctx context.Context, filter model.ExpenseFilter) ([]model.Expense, int64, error) {
	var expenses []model.Expense
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Expense{}).Preload("User")

	if filter.TenantID != 0 {
		query = query.Where("expenses.tenant_id = ?", filter.TenantID)
	}

	if filter.UserID != 0 {
		query = query.Where("expenses.user_id = ?", filter.UserID)
	}

	if filter.Status != "" {
		query = query.Where("expenses.status = ?", filter.Status)
	}

	if filter.Search != "" {
		searchTerm := "%" + filter.Search + "%"
		query = query.Joins("User").Where("expenses.id::text LIKE ? OR \"User\".name ILIKE ?", searchTerm, searchTerm)
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

	err := query.Order("expenses.created_at DESC").Find(&expenses).Error
	return expenses, total, err
}

func (r *expenseRepository) FindByID(ctx context.Context, id uint) (*model.Expense, error) {
	var expense model.Expense
	err := r.db.WithContext(ctx).Preload("User").First(&expense, id).Error
	if err != nil {
		return nil, err
	}
	return &expense, nil
}

func (r *expenseRepository) Create(ctx context.Context, expense *model.Expense) error {
	return r.db.WithContext(ctx).Create(expense).Error
}

func (r *expenseRepository) Update(ctx context.Context, expense *model.Expense) error {
	return r.db.WithContext(ctx).Save(expense).Error
}

func (r *expenseRepository) GetSummary(ctx context.Context, tenantID uint) (float64, float64, string, float64, error) {
	var pendingAmount float64
	var approvedThisMonth float64

	// Pending Amount
	r.db.WithContext(ctx).Model(&model.Expense{}).
		Where("tenant_id = ? AND status = ?", tenantID, model.ExpenseStatusPending).
		Select("COALESCE(SUM(amount), 0)").Scan(&pendingAmount)

	// Approved This Month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	r.db.WithContext(ctx).Model(&model.Expense{}).
		Where("tenant_id = ? AND status = ? AND date >= ?", tenantID, model.ExpenseStatusApproved, startOfMonth).
		Select("COALESCE(SUM(amount), 0)").Scan(&approvedThisMonth)

	// Top Category
	type CatStats struct {
		Category model.ExpenseCategory
		Total    float64
	}
	var stats []CatStats
	var totalApproved float64

	r.db.WithContext(ctx).Model(&model.Expense{}).
		Where("tenant_id = ? AND status = ?", tenantID, model.ExpenseStatusApproved).
		Select("category, SUM(amount) as total").
		Group("category").
		Order("total DESC").
		Scan(&stats)

	r.db.WithContext(ctx).Model(&model.Expense{}).
		Where("tenant_id = ? AND status = ?", tenantID, model.ExpenseStatusApproved).
		Select("COALESCE(SUM(amount), 0)").Scan(&totalApproved)

	topCat := "None"
	topPct := 0.0
	if len(stats) > 0 && totalApproved > 0 {
		topCat = string(stats[0].Category)
		topPct = (stats[0].Total / totalApproved) * 100
	}

	return pendingAmount, approvedThisMonth, topCat, topPct, nil
}
