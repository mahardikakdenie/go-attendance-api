package service

import (
	"context"
	"errors"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
)

type TenantRoleService interface {
	ListRoles(ctx context.Context, tenantID uint) ([]model.Role, error)
	CreateRole(ctx context.Context, tenantID uint, req CreateRoleRequest) (*model.Role, error)
	UpdateRole(ctx context.Context, tenantID uint, roleID uint, req UpdateRoleRequest) (*model.Role, error)
	DeleteRole(ctx context.Context, tenantID uint, roleID uint) error
	GetHierarchy(ctx context.Context, roleID uint) ([]model.RoleHierarchy, error)
	SaveHierarchy(ctx context.Context, tenantID uint, parentRoleID uint, childRoleIDs []uint) error
}

type CreateRoleRequest struct {
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description"`
	BaseRole    model.BaseRole `json:"base_role" binding:"required"`
	Permissions []string       `json:"permissions"`
}

type UpdateRoleRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Permissions []string `json:"permissions"`
}

type tenantRoleService struct {
	roleRepo      repository.RoleRepository
	permissionRepo repository.PermissionRepository
	hierarchyRepo repository.RoleHierarchyRepository
}

func NewTenantRoleService(
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	hierarchyRepo repository.RoleHierarchyRepository,
) TenantRoleService {
	return &tenantRoleService{
		roleRepo:      roleRepo,
		permissionRepo: permissionRepo,
		hierarchyRepo: hierarchyRepo,
	}
}

func (s *tenantRoleService) ListRoles(ctx context.Context, tenantID uint) ([]model.Role, error) {
	return s.roleRepo.FindByTenantID(ctx, tenantID)
}

func (s *tenantRoleService) CreateRole(ctx context.Context, tenantID uint, req CreateRoleRequest) (*model.Role, error) {
	role := &model.Role{
		TenantID:    &tenantID,
		Name:        req.Name,
		Description: req.Description,
		BaseRole:    req.BaseRole,
		IsSystem:    false,
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return nil, err
	}

	if len(req.Permissions) > 0 {
		if err := s.roleRepo.UpdatePermissions(ctx, role.ID, req.Permissions); err != nil {
			return nil, err
		}
	}

	return s.roleRepo.FindByID(ctx, role.ID)
}

func (s *tenantRoleService) UpdateRole(ctx context.Context, tenantID uint, roleID uint, req UpdateRoleRequest) (*model.Role, error) {
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return nil, err
	}

	if role.TenantID == nil || *role.TenantID != tenantID {
		return nil, errors.New("forbidden: not your tenant role")
	}

	if req.Name != "" {
		role.Name = req.Name
	}
	if req.Description != "" {
		role.Description = req.Description
	}

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return nil, err
	}

	if req.Permissions != nil {
		if err := s.roleRepo.UpdatePermissions(ctx, role.ID, req.Permissions); err != nil {
			return nil, err
		}
	}

	return s.roleRepo.FindByID(ctx, role.ID)
}

func (s *tenantRoleService) DeleteRole(ctx context.Context, tenantID uint, roleID uint) error {
	role, err := s.roleRepo.FindByID(ctx, roleID)
	if err != nil {
		return err
	}

	if role.IsSystem {
		return errors.New("forbidden: cannot delete system role")
	}

	if role.TenantID == nil || *role.TenantID != tenantID {
		return errors.New("forbidden: not your tenant role")
	}

	return s.roleRepo.Delete(ctx, roleID)
}

func (s *tenantRoleService) GetHierarchy(ctx context.Context, roleID uint) ([]model.RoleHierarchy, error) {
	return s.hierarchyRepo.FindByParentID(ctx, roleID)
}

func (s *tenantRoleService) SaveHierarchy(ctx context.Context, tenantID uint, parentRoleID uint, childRoleIDs []uint) error {
	// Verify parent role belongs to tenant
	parentRole, err := s.roleRepo.FindByID(ctx, parentRoleID)
	if err != nil {
		return err
	}
	if parentRole.TenantID != nil && *parentRole.TenantID != tenantID {
		return errors.New("forbidden: parent role not in your tenant")
	}

	if err := s.hierarchyRepo.DeleteByParentID(ctx, parentRoleID); err != nil {
		return err
	}
	return s.addHierarchies(ctx, tenantID, parentRoleID, childRoleIDs)
}

func (s *tenantRoleService) addHierarchies(ctx context.Context, tenantID uint, parentRoleID uint, childRoleIDs []uint) error {
	for _, childID := range childRoleIDs {
		rh := &model.RoleHierarchy{
			TenantID:     tenantID,
			ParentRoleID: parentRoleID,
			ChildRoleID:  childID,
		}
		if err := s.hierarchyRepo.Create(ctx, rh); err != nil {
			return err
		}
	}
	return nil
}
