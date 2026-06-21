package repository

import (
	"context"
	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type MenuRepository interface {
	FindAll(ctx context.Context) ([]model.Menu, error)
	FindAllWithRoles(ctx context.Context) ([]model.Menu, error)
	Create(ctx context.Context, menu *model.Menu) error
	Update(ctx context.Context, menu *model.Menu) error
	UpdateWithRoles(ctx context.Context, menu *model.Menu, roleIDs []uint) error
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
	err := r.db.WithContext(ctx).Order("parent_id ASC, sort_order ASC").Find(&menus).Error
	return menus, err
}

func (r *menuRepository) FindAllWithRoles(ctx context.Context) ([]model.Menu, error) {
	var menus []model.Menu
	err := r.db.WithContext(ctx).
		Preload("Roles").
		Order("parent_id ASC, sort_order ASC").
		Find(&menus).Error
	return menus, err
}

func (r *menuRepository) Create(ctx context.Context, menu *model.Menu) error {
	return r.db.WithContext(ctx).Create(menu).Error
}

func (r *menuRepository) Update(ctx context.Context, menu *model.Menu) error {
	return r.db.WithContext(ctx).Save(menu).Error
}

func (r *menuRepository) UpdateWithRoles(ctx context.Context, menu *model.Menu, roleIDs []uint) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Update basic fields
		if err := tx.Save(menu).Error; err != nil {
			return err
		}

		// Update many-to-many relationship
		var roles []model.Role
		if len(roleIDs) > 0 {
			if err := tx.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
				return err
			}
		}

		return tx.Model(menu).Association("Roles").Replace(roles)
	})
}

func (r *menuRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Menu{}, id).Error
}
