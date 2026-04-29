package repository

import (
	"context"
	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"time"

	"gorm.io/gorm"
)

type SuperadminRepository interface {
	GetOwnersWithStats(ctx context.Context, limit, offset int) ([]modelDto.OwnerWithStatsResponse, int64, error)
	GetPlatformAccounts(ctx context.Context, search string, limit, offset int) ([]model.User, int64, error)
	GetAnalyticsDashboard(ctx context.Context, period string) (*modelDto.AnalyticsDashboardResponse, error)
}

type superadminRepository struct {
	db *gorm.DB
}

func NewSuperadminRepository(db *gorm.DB) SuperadminRepository {
	return &superadminRepository{db: db}
}

func (r *superadminRepository) GetOwnersWithStats(ctx context.Context, limit, offset int) ([]modelDto.OwnerWithStatsResponse, int64, error) {
	var results []modelDto.OwnerWithStatsResponse
	var total int64

	// Count total owners (BaseRole = ADMIN)
	err := r.db.WithContext(ctx).Model(&model.User{}).
		Joins("JOIN roles ON users.role_id = roles.id").
		Where("roles.base_role = ?", model.BaseRoleAdmin).
		Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	// Fetch owners with pagination
	// We use subqueries for counts to keep it in one query, but only for the paginated results.
	// Optimization: This query only runs for the current page of users.
	query := `
		SELECT 
			u.id, u.name, u.email, u.created_at, u.tenant_id,
			t.name as tenant_name, t.code as tenant_code, t.plan as tenant_plan,
			(SELECT COUNT(*) FROM users WHERE tenant_id = u.tenant_id) as employee_count,
			(SELECT COUNT(*) FROM attendances WHERE tenant_id = u.tenant_id) as attendance_count,
			(SELECT COUNT(*) FROM leaves WHERE tenant_id = u.tenant_id) as leave_count,
			(SELECT COUNT(*) FROM overtimes WHERE tenant_id = u.tenant_id) as overtime_count,
			(SELECT COUNT(*) FROM payrolls WHERE tenant_id = u.tenant_id) as payroll_count,
			(SELECT COUNT(*) FROM expenses WHERE tenant_id = u.tenant_id) as expense_count
		FROM users u
		JOIN roles r ON u.role_id = r.id
		JOIN tenants t ON u.tenant_id = t.id
		WHERE r.base_role = ?
		ORDER BY u.created_at DESC
		LIMIT ? OFFSET ?
	`

	err = r.db.WithContext(ctx).Raw(query, model.BaseRoleAdmin, limit, offset).Scan(&results).Error
	if err != nil {
		return nil, 0, err
	}

	return results, total, nil
}

func (r *superadminRepository) GetPlatformAccounts(ctx context.Context, search string, limit, offset int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.WithContext(ctx).Model(&model.User{}).
		Joins("JOIN roles ON users.role_id = roles.id").
		Where("roles.base_role IN ?", []model.BaseRole{model.BaseRoleSuperAdmin, model.BaseRoleSupport, model.BaseRoleEngineer})

	if search != "" {
		query = query.Where("users.name ILIKE ? OR users.email ILIKE ?", "%"+search+"%", "%"+search+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = query.Preload("Role").Preload("Role.Permissions").
		Order("users.created_at DESC").
		Limit(limit).Offset(offset).
		Find(&users).Error

	return users, total, err
}

func (r *superadminRepository) GetAnalyticsDashboard(ctx context.Context, period string) (*modelDto.AnalyticsDashboardResponse, error) {
	var response modelDto.AnalyticsDashboardResponse
	now := time.Now()
	thirtyDaysAgo := now.AddDate(0, 0, -30)

	// 1. KPIs
	// Total Tenants
	var totalTenants, totalTenantsPrev int64
	r.db.WithContext(ctx).Model(&model.Tenant{}).Count(&totalTenants)
	r.db.WithContext(ctx).Model(&model.Tenant{}).Where("created_at < ?", thirtyDaysAgo).Count(&totalTenantsPrev)
	response.KPIs.TotalTenants = modelDto.KPIData{
		Value:     totalTenants,
		GrowthPct: calculateGrowth(totalTenants, totalTenantsPrev),
	}

	// Total Users
	var totalUsers, totalUsersPrev int64
	r.db.WithContext(ctx).Model(&model.User{}).Count(&totalUsers)
	r.db.WithContext(ctx).Model(&model.User{}).Where("created_at < ?", thirtyDaysAgo).Count(&totalUsersPrev)
	response.KPIs.TotalUsers = modelDto.KPIData{
		Value:     totalUsers,
		GrowthPct: calculateGrowth(totalUsers, totalUsersPrev),
	}

	// Active Subscriptions
	var activeSubs, activeSubsPrev int64
	r.db.WithContext(ctx).Model(&model.Subscription{}).Where("status = ?", model.SubscriptionStatusActive).Count(&activeSubs)
	r.db.WithContext(ctx).Model(&model.Subscription{}).Where("status = ? AND created_at < ?", model.SubscriptionStatusActive, thirtyDaysAgo).Count(&activeSubsPrev)
	response.KPIs.ActiveSubscriptions = modelDto.KPIData{
		Value:     activeSubs,
		GrowthPct: calculateGrowth(activeSubs, activeSubsPrev),
	}

	// Monthly Growth (Tenant acquisition last 30 days vs 30-60 days ago)
	var last30DaysTenants, prev30DaysTenants int64
	sixtyDaysAgo := now.AddDate(0, 0, -60)
	r.db.WithContext(ctx).Model(&model.Tenant{}).Where("created_at BETWEEN ? AND ?", thirtyDaysAgo, now).Count(&last30DaysTenants)
	r.db.WithContext(ctx).Model(&model.Tenant{}).Where("created_at BETWEEN ? AND ?", sixtyDaysAgo, thirtyDaysAgo).Count(&prev30DaysTenants)
	growthRate := calculateGrowth(last30DaysTenants, prev30DaysTenants)

	// Assuming Value is the actual count of new tenants and GrowthPct is the percentage change
	response.KPIs.MonthlyGrowth = modelDto.KPIData{
		Value:     last30DaysTenants,
		GrowthPct: growthRate,
	}

	// 2. Tenant Growth Trend
	var monthsLimit int
	if period == "last_6_months" {
		monthsLimit = 6
	} else {
		monthsLimit = 12 // this_year default
	}

	// Aggregate tenants by month using SQL
	var growthResults []struct {
		Month string
		Count int64
	}

	// SQL for PostgreSQL
	queryTrend := `
		SELECT to_char(created_at, 'Mon') as month, COUNT(*) as count, to_char(created_at, 'MM') as month_num
		FROM tenants
		WHERE created_at >= ?
		GROUP BY month, month_num
		ORDER BY month_num ASC
	`
	startDate := now.AddDate(0, -monthsLimit, 0)
	r.db.WithContext(ctx).Raw(queryTrend, startDate).Scan(&growthResults)

	response.GrowthChart.Labels = make([]string, 0, len(growthResults))
	response.GrowthChart.Data = make([]int64, 0, len(growthResults))

	for _, res := range growthResults {
		response.GrowthChart.Labels = append(response.GrowthChart.Labels, res.Month)
		response.GrowthChart.Data = append(response.GrowthChart.Data, res.Count)
	}

	// 3. Tenant Status Distribution
	var activeCount, trialCount, suspendedCount int64

	// Count Suspended Tenants
	r.db.WithContext(ctx).Model(&model.Tenant{}).Where("is_suspended = ?", true).Count(&suspendedCount)

	// Count Active Subscriptions (for non-suspended tenants)
	r.db.WithContext(ctx).Model(&model.Subscription{}).
		Joins("JOIN tenants ON subscriptions.tenant_id = tenants.id").
		Where("LOWER(subscriptions.status) = ? AND tenants.is_suspended = ?", "active", false).
		Count(&activeCount)

	// Count Trial Subscriptions (for non-suspended tenants)
	r.db.WithContext(ctx).Model(&model.Subscription{}).
		Joins("JOIN tenants ON subscriptions.tenant_id = tenants.id").
		Where("LOWER(subscriptions.status) = ? AND tenants.is_suspended = ?", "trial", false).
		Count(&trialCount)

	response.TenantStatus = []modelDto.TenantStatusData{
		{Label: "Active", Value: activeCount, Color: "#10b981"},
		{Label: "Suspended", Value: suspendedCount, Color: "#f43f5e"},
		{Label: "Trial", Value: trialCount, Color: "#f59e0b"},
	}

	return &response, nil
}

func calculateGrowth(current, previous int64) float64 {
	if previous == 0 {
		if current > 0 {
			return 100.0
		}
		return 0.0
	}
	return (float64(current-previous) / float64(previous)) * 100.0
}
