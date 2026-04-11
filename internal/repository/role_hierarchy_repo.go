package repository

import (
	"context"
	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type RoleHierarchyRepository interface {
	FindByParentID(ctx context.Context, parentID uint) ([]model.RoleHierarchy, error)
	DeleteByParentID(ctx context.Context, parentID uint) error
	Create(ctx context.Context, rh *model.RoleHierarchy) error
	GetChildRoleIDs(ctx context.Context, parentRoleID uint) ([]uint, error)
	GetAllDescendantRoleIDs(ctx context.Context, parentRoleID uint) ([]uint, error)
}

type roleHierarchyRepository struct {
	db *gorm.DB
}

func NewRoleHierarchyRepository(db *gorm.DB) RoleHierarchyRepository {
	return &roleHierarchyRepository{
		db: db,
	}
}

func (r *roleHierarchyRepository) FindByParentID(ctx context.Context, parentID uint) ([]model.RoleHierarchy, error) {
	var results []model.RoleHierarchy
	err := r.db.WithContext(ctx).
		Where("parent_role_id = ?", parentID).
		Preload("ChildRole").
		Find(&results).Error
	return results, err
}

func (r *roleHierarchyRepository) DeleteByParentID(ctx context.Context, parentID uint) error {
	return r.db.WithContext(ctx).Where("parent_role_id = ?", parentID).Delete(&model.RoleHierarchy{}).Error
}

func (r *roleHierarchyRepository) Create(ctx context.Context, rh *model.RoleHierarchy) error {
	return r.db.WithContext(ctx).Create(rh).Error
}

func (r *roleHierarchyRepository) GetChildRoleIDs(ctx context.Context, parentRoleID uint) ([]uint, error) {
	var roleIDs []uint
	err := r.db.WithContext(ctx).Model(&model.RoleHierarchy{}).
		Where("parent_role_id = ?", parentRoleID).
		Pluck("child_role_id", &roleIDs).Error
	return roleIDs, err
}

func (r *roleHierarchyRepository) GetAllDescendantRoleIDs(ctx context.Context, parentRoleID uint) ([]uint, error) {
	var allDescendants []uint
	visited := make(map[uint]bool)
	
	var traverse func(roleID uint) error
	traverse = func(roleID uint) error {
		if visited[roleID] {
			return nil
		}
		visited[roleID] = true
		
		children, err := r.GetChildRoleIDs(ctx, roleID)
		if err != nil {
			return err
		}
		
		for _, childID := range children {
			allDescendants = append(allDescendants, childID)
			if err := traverse(childID); err != nil {
				return err
			}
		}
		return nil
	}
	
	err := traverse(parentRoleID)
	return allDescendants, err
}
