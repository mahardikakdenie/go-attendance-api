package repository

import (
	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type UserRepository interface {
	FindAll(filter model.UserFilter) ([]model.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) FindAll(filter model.UserFilter) ([]model.User, error) {
	var users []model.User

	query := r.db.Model(&model.User{})

	if filter.Name != "" {
		query = query.Where("name ILIKE ?", "%"+filter.Name+"%")
	}

	if filter.Email != "" {
		query = query.Where("email = ?", filter.Email)
	}

	if filter.OrderBy != "" {
		sortDir := "ASC"
		if filter.Sort == "desc" || filter.Sort == "DESC" {
			sortDir = "DESC"
		}
		query = query.Order(filter.OrderBy + " " + sortDir)
	}

	err := query.Find(&users).Error

	return users, err
}
