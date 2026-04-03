package repository

import (
	"context"
	"errors"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

var ErrUserNotFound = errors.New("user not found")

type UserRepository interface {
	FindAll(ctx context.Context, filter model.UserFilter) ([]model.User, int64, error)
	FindByID(ctx context.Context, id uint) (*model.User, error)
	GetMe(ctx context.Context, userID uint) (*model.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

// ======================
// FIND BY ID (FIX ERROR LO 🔥)
// ======================
func (r *userRepository) FindByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User

	err := r.db.WithContext(ctx).
		First(&user, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// ======================
// FIND ALL (FILTER + PAGINATION)
// ======================
func (r *userRepository) FindAll(ctx context.Context, filter model.UserFilter) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.WithContext(ctx).Model(&model.User{})

	// 🔥 FILTER
	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}

	if filter.Email != "" {
		query = query.Where("email = ?", filter.Email)
	}

	if filter.Role != "" {
		query = query.Where("role = ?", filter.Role)
	}

	if filter.TenantID != 0 {
		query = query.Where("tenant_id = ?", filter.TenantID)
	}

	// 🔥 COUNT TOTAL
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 🔥 SORTING
	if filter.OrderBy != "" {
		sortDir := "ASC"
		if filter.Sort == "desc" || filter.Sort == "DESC" {
			sortDir = "DESC"
		}
		query = query.Order(filter.OrderBy + " " + sortDir)
	}

	// 🔥 PAGINATION
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}

	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	// 🔥 EXECUTE
	if err := query.Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ======================
// GET ME
// ======================
func (r *userRepository) GetMe(ctx context.Context, userID uint) (*model.User, error) {
	var user model.User

	err := r.db.WithContext(ctx).
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
