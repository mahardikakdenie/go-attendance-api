package repository

import (
	"context"
	"go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type SuperadminRepository interface {
	GetOwnersWithStats(ctx context.Context, limit, offset int) ([]modelDto.OwnerWithStatsResponse, int64, error)
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
			t.name as tenant_name, t.code as tenant_code,
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
