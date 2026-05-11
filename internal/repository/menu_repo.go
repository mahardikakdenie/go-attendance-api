package repository

import (
	"context"
	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type MenuRepository interface {
	FindAll(ctx context.Context) ([]model.Menu, error)
	Create(ctx context.Context, menu *model.Menu) error
	Update(ctx context.Context, menu *model.Menu) error
	Delete(ctx context.Context, id uint) error
}

type menuRepository struct {
	db *gorm.DB
}

func NewMenuRepository(db *gorm.DB) MenuRepository {
	return &menuRepository{db: db}
}

func (r *menuRepository) FindAll(ctx context.Context) ([]model.Menu, error) {
	var menus []model.Menu
	// Preload only top-level children. Service will handle recursion if needed,
	// but usually we just fetch all and build tree in memory for efficiency.
	err := r.db.WithContext(ctx).Order("parent_id ASC, sort_order ASC").Find(&menus).Error
	return menus, err
}

func (r *menuRepository) Create(ctx context.Context, menu *model.Menu) error {
	return r.db.WithContext(ctx).Create(menu).Error
}

func (r *menuRepository) Update(ctx context.Context, menu *model.Menu) error {
	return r.db.WithContext(ctx).Save(menu).Error
}

func (r *menuRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Menu{}, id).Error
}
