package repository

import (
	"context"
	"errors"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/utils"

	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	FindAll(ctx context.Context, filter model.UserFilter, includes []string) ([]model.User, int64, error)
	FindByID(ctx context.Context, id uint, includes []string) (*model.User, error)
	GetMe(ctx context.Context, userID uint, includes []string) (*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Create(ctx context.Context, user *model.User) error
	CountByTenantID(ctx context.Context, tenantID uint) (int64, error)
	FindTenantByID(ctx context.Context, tenantID uint) (*model.Tenant, error)
	UpdateQuota(ctx context.Context, userID uint, quota float64) error
	DecreaseQuota(ctx context.Context, userID uint, amount float64) error
	Transaction(ctx context.Context, fn func(repo UserRepository) error) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Transaction(ctx context.Context, fn func(repo UserRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(NewUserRepository(tx))
	})
}

var userPreloadMap = map[string]string{
	"tenant":                  "Tenant",
	"tenant.tenant_settings": "Tenant.TenantSettings",
	"tenant_setting":          "Tenant.TenantSettings",
	"attendances":             "Attendances",
	"attendances.user":        "Attendances.User",
	"role":                    "Role",
	"role.permissions":        "Role.Permissions",
	"position":                "Position",
	"recent_activities":       "RecentActivities",
	"manager":                 "Manager",
	"delegate":                "Delegate",
}

func (r *userRepository) FindByID(ctx context.Context, id uint, includes []string) (*model.User, error) {
	var user model.User

	query := r.db.WithContext(ctx)
	query = utils.ApplyPreloads(query, includes, userPreloadMap)

	err := query.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) FindAll(
	ctx context.Context,
	filter model.UserFilter,
	includes []string,
) ([]model.User, int64, error) {

	var users []model.User
	var total int64

	query := r.db.WithContext(ctx).Model(&model.User{})
	query = utils.ApplyPreloads(query, includes, userPreloadMap)

	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}

	if filter.Email != "" {
		query = query.Where("email = ?", filter.Email)
	}

	if filter.RoleID != 0 {
		query = query.Where("role_id = ?", filter.RoleID)
	}

	if len(filter.AllowedRoleIDs) > 0 {
		query = query.Where("role_id IN ?", filter.AllowedRoleIDs)
	}

	if filter.TenantID != 0 {
		query = query.Where("users.tenant_id = ?", filter.TenantID)
	}

	if filter.EmployeeID != "" {
		query = query.Where("users.employee_id ILIKE ?", "%"+filter.EmployeeID+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if filter.OrderBy != "" {
		sortDir := "ASC"
		if filter.Sort == "desc" || filter.Sort == "DESC" {
			sortDir = "DESC"
		}
		query = query.Order(filter.OrderBy + " " + sortDir)
	}

	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

func (r *userRepository) GetMe(ctx context.Context, userID uint, includes []string) (*model.User, error) {
	var user model.User

	query := r.db.WithContext(ctx)
	query = utils.ApplyPreloads(query, includes, userPreloadMap)

	err := query.
		Where("id = ?", userID).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	if user.ID == 0 {
		return errors.New("invalid user id")
	}

	// pastikan user ada
	var existing model.User
	if err := r.db.WithContext(ctx).
		First(&existing, user.ID).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return err
	}

	// update hanya field tertentu (biar aman)
	updateData := map[string]interface{}{
		"name":         user.Name,
		"email":        user.Email,
		"role_id":      user.RoleID,
		"tenant_id":    user.TenantID,
		"media_url":    user.MediaUrl,
		"employee_id":  user.EmployeeID,
		"department":   user.Department,
		"address":      user.Address,
		"phone_number": user.PhoneNumber,
		"updated_at":   user.UpdatedAt,
	}

	if err := r.db.WithContext(ctx).
		Model(&existing).
		Updates(updateData).Error; err != nil {
		return err
	}

	return nil
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) CountByTenantID(ctx context.Context, tenantID uint) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Where("tenant_id = ?", tenantID).Count(&count).Error
	return count, err
}

func (r *userRepository) FindTenantByID(ctx context.Context, tenantID uint) (*model.Tenant, error) {
	var tenant model.Tenant
	err := r.db.WithContext(ctx).Preload("TenantSettings").First(&tenant, tenantID).Error
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (r *userRepository) UpdateQuota(ctx context.Context, userID uint, quota float64) error {
	return r.db.WithContext(ctx).Model(&model.User{}).Where("id = ?", userID).Update("expense_quota", quota).Error
}

func (r *userRepository) DecreaseQuota(ctx context.Context, userID uint, amount float64) error {
	return r.db.WithContext(ctx).Model(&model.User{}).
		Where("id = ?", userID).
		Update("expense_quota", gorm.Expr("expense_quota - ?", amount)).Error
}
