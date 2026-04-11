package repository

import (
	"context"
	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type RoleRepository interface {
	FindByName(ctx context.Context, name string) (*model.Role, error)
	FindByID(ctx context.Context, id uint) (*model.Role, error)
	FindByTenantID(ctx context.Context, tenantID uint) ([]model.Role, error)
	Create(ctx context.Context, role *model.Role) error
	Update(ctx context.Context, role *model.Role) error
	Delete(ctx context.Context, id uint) error
	UpdatePermissions(ctx context.Context, roleID uint, permissionIDs []string) error
}

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *gorm.DB) RoleRepository {
	return &roleRepository{
		db: db,
	}
}

func (r *roleRepository) FindByName(ctx context.Context, name string) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByID(ctx context.Context, id uint) (*model.Role, error) {
	var role model.Role
	err := r.db.WithContext(ctx).Preload("Permissions").First(&role, id).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *roleRepository) FindByTenantID(ctx context.Context, tenantID uint) ([]model.Role, error) {
	var roles []model.Role
	// Fetch System Roles (tenant_id is NULL) and Custom Tenant Roles
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? OR tenant_id IS NULL", tenantID).
		Preload("Permissions").
		Find(&roles).Error
	return roles, err
}

func (r *roleRepository) Create(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Create(role).Error
}

func (r *roleRepository) Update(ctx context.Context, role *model.Role) error {
	return r.db.WithContext(ctx).Save(role).Error
}

func (r *roleRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Role{}, id).Error
}

func (r *roleRepository) UpdatePermissions(ctx context.Context, roleID uint, permissionIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Delete existing permissions
		if err := tx.Where("role_id = ?", roleID).Delete(&model.RolePermission{}).Error; err != nil {
			return err
		}

		// Add new permissions
		for _, pID := range permissionIDs {
			rp := model.RolePermission{
				RoleID:       roleID,
				PermissionID: pID,
			}
			if err := tx.Create(&rp).Error; err != nil {
				return err
			}
		}
		return nil
	})
}
