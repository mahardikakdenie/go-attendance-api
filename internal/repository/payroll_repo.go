package repository

import (
	"context"
	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type PayrollRepository interface {
	Save(ctx context.Context, payroll *model.Payroll) error
	Update(ctx context.Context, payroll *model.Payroll) error
	FindByID(ctx context.Context, id uint) (*model.Payroll, error)
	FindByUserPeriod(ctx context.Context, userID uint, period string) (*model.Payroll, error)
	FindAll(ctx context.Context, tenantID uint, period string, search string, limit, offset int) ([]model.Payroll, int64, error)
	GetSummary(ctx context.Context, tenantID uint, period string) (model.PayrollSummary, error)
	DeleteByPeriod(ctx context.Context, tenantID uint, period string) error
	FindAllByUser(ctx context.Context, userID uint, status string) ([]model.Payroll, error)
}

type payrollRepository struct {
	db *gorm.DB
}

func NewPayrollRepository(db *gorm.DB) PayrollRepository {
	return &payrollRepository{db: db}
}

func (r *payrollRepository) Save(ctx context.Context, payroll *model.Payroll) error {
	return r.db.WithContext(ctx).Create(payroll).Error
}

func (r *payrollRepository) Update(ctx context.Context, payroll *model.Payroll) error {
	return r.db.WithContext(ctx).Save(payroll).Error
}

func (r *payrollRepository) FindByID(ctx context.Context, id uint) (*model.Payroll, error) {
	var payroll model.Payroll
	err := r.db.WithContext(ctx).Preload("User").First(&payroll, id).Error
	if err != nil {
		return nil, err
	}
	return &payroll, nil
}

func (r *payrollRepository) FindByUserPeriod(ctx context.Context, userID uint, period string) (*model.Payroll, error) {
	var payroll model.Payroll
	err := r.db.WithContext(ctx).Where("user_id = ? AND period = ?", userID, period).First(&payroll).Error
	if err != nil {
		return nil, err
	}
	return &payroll, nil
}

func (r *payrollRepository) FindAll(ctx context.Context, tenantID uint, period string, search string, limit, offset int) ([]model.Payroll, int64, error) {
	var payrolls []model.Payroll
	var total int64

	query := r.db.WithContext(ctx).Model(&model.Payroll{}).
		Preload("User.Position").
		Where("tenant_id = ? AND period = ?", tenantID, period)

	if search != "" {
		query = query.Joins("JOIN users ON users.id = payrolls.user_id").
			Where("users.name ILIKE ? OR users.employee_id ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	query.Count(&total)

	err := query.Limit(limit).Offset(offset).Find(&payrolls).Error
	return payrolls, total, err
}

func (r *payrollRepository) GetSummary(ctx context.Context, tenantID uint, period string) (model.PayrollSummary, error) {
	var summary model.PayrollSummary

	// Total Net Payout, Tax, and BPJS
	err := r.db.WithContext(ctx).Model(&model.Payroll{}).
		Select("SUM(net_salary) as total_net_payout, SUM(pph21_amount) as total_tax_liability, SUM(bpjs_health_employee + bpjs_jht_employee) as total_bpjs_provision").
		Where("tenant_id = ? AND period = ?", tenantID, period).
		Scan(&summary).Error

	// Simplified logic for sync rate and diff for demo purposes
	summary.AttendanceSyncRate = 100.0 // Default
	summary.PayoutDiffPercentage = 0.0

	return summary, err
}

func (r *payrollRepository) DeleteByPeriod(ctx context.Context, tenantID uint, period string) error {
	return r.db.WithContext(ctx).Where("tenant_id = ? AND period = ? AND status = ?", tenantID, period, model.PayrollStatusDraft).Delete(&model.Payroll{}).Error
}

func (r *payrollRepository) FindAllByUser(ctx context.Context, userID uint, status string) ([]model.Payroll, error) {
	var payrolls []model.Payroll
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	err := query.Order("period DESC").Find(&payrolls).Error
	return payrolls, err
}
