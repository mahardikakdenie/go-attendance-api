package service

import (
	"context"
	"errors"
	"fmt"
	"go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/utils"

	"golang.org/x/crypto/bcrypt"
)

type SuperadminService interface {
	GetOwnersWithStats(ctx context.Context, limit, offset int) ([]modelDto.OwnerWithStatsResponse, int64, error)
	GetPlatformAccounts(ctx context.Context, search string, limit, offset int) ([]model.UserResponse, int64, error)
	CreatePlatformAccount(ctx context.Context, req model.CreateUserRequest) (model.UserResponse, error)
	UpdatePlatformAccount(ctx context.Context, id uint, req model.CreateUserRequest) (model.UserResponse, error)
	TogglePlatformAccountStatus(ctx context.Context, id uint, isActive bool) error

	// System Role Management
	ListSystemRoles(ctx context.Context) ([]model.Role, error)
	ListAllPermissions(ctx context.Context) ([]model.Permission, error)
	CreateSystemRole(ctx context.Context, req modelDto.CreateSystemRoleRequest, performerID uint) (model.Role, error)
	UpdateSystemRole(ctx context.Context, id uint, req modelDto.CreateSystemRoleRequest, performerID uint) (model.Role, error)
	DeleteSystemRole(ctx context.Context, id uint, performerID uint) error
}

type superadminService struct {
	repo           repository.SuperadminRepository
	userRepo       repository.UserRepository
	roleRepo       repository.RoleRepository
	permissionRepo repository.PermissionRepository
	activityRepo   repository.RecentActivityRepository
}

func NewSuperadminService(
	repo repository.SuperadminRepository,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	permissionRepo repository.PermissionRepository,
	activityRepo repository.RecentActivityRepository,
) SuperadminService {
	return &superadminService{
		repo:           repo,
		userRepo:       userRepo,
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		activityRepo:   activityRepo,
	}
}

func (s *superadminService) GetOwnersWithStats(ctx context.Context, limit, offset int) ([]modelDto.OwnerWithStatsResponse, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.GetOwnersWithStats(ctx, limit, offset)
}

func (s *superadminService) GetPlatformAccounts(ctx context.Context, search string, limit, offset int) ([]model.UserResponse, int64, error) {
	if limit <= 0 {
		limit = 10
	}

	users, total, err := s.repo.GetPlatformAccounts(ctx, search, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var responses []model.UserResponse
	for _, user := range users {
		responses = append(responses, model.UserResponse{
			ID:        user.ID,
			Name:      user.Name,
			Email:     user.Email,
			TenantID:  user.TenantID,
			IsActive:  user.IsActive,
			CreatedAt: user.CreatedAt,
			Role: &model.RoleResponse{
				ID:       user.Role.ID,
				Name:     user.Role.Name,
				BaseRole: user.Role.BaseRole,
			},
			BaseRole: user.Role.BaseRole,
		})
	}

	return responses, total, nil
}

func (s *superadminService) CreatePlatformAccount(ctx context.Context, req model.CreateUserRequest) (model.UserResponse, error) {
	// Platform accounts are always tenant 1 (HQ)
	req.TenantID = 1

	// Validate Role is a system role (SUPERADMIN, SUPPORT, ENGINEER)
	role, err := s.roleRepo.FindByID(ctx, req.RoleID)
	if err != nil || role == nil {
		return model.UserResponse{}, errors.New("invalid role")
	}

	if role.BaseRole != model.BaseRoleSuperAdmin && role.BaseRole != model.BaseRoleSupport && role.BaseRole != model.BaseRoleEngineer {
		return model.UserResponse{}, errors.New("role must be a system role (SUPERADMIN, SUPPORT, or ENGINEER)")
	}

	password := req.Password
	if password == "" {
		password = utils.GenerateRandomString(12)
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	user := &model.User{
		Name:               req.Name,
		Email:              req.Email,
		Password:           string(hashedPassword),
		RoleID:             req.RoleID,
		TenantID:           1,
		IsSystemCreated:    true,
		MustChangePassword: true,
		IsActive:           true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return model.UserResponse{}, err
	}

	// Audit Log
	_ = s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: 1, // System/Root Admin ID placeholder
		Title:  "Platform Administration",
		Action: fmt.Sprintf("Created platform account: %s (%s)", user.Name, role.BaseRole),
		Status: "success",
	})

	// Send Email
	emailHtml := utils.GetWelcomeEmailTemplate(user.Name, user.Email, password, "Attendance System HQ", "")
	subject := "Platform Administrator Account Created"
	go func() {
		_ = utils.SendEmail([]string{user.Email}, subject, emailHtml)
	}()

	return model.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		TenantID:  user.TenantID,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		BaseRole:  role.BaseRole,
	}, nil
}

func (s *superadminService) UpdatePlatformAccount(ctx context.Context, id uint, req model.CreateUserRequest) (model.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, id, []string{"role"})
	if err != nil {
		return model.UserResponse{}, errors.New("account not found")
	}

	// Protection: Cannot update ID 1 (Primary Root) via this API easily to avoid self-lockout
	if id == 1 {
		return model.UserResponse{}, errors.New("primary root admin cannot be modified via this API")
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	if req.RoleID != 0 {
		role, err := s.roleRepo.FindByID(ctx, req.RoleID)
		if err == nil && role != nil {
			if role.BaseRole == model.BaseRoleSuperAdmin || role.BaseRole == model.BaseRoleSupport || role.BaseRole == model.BaseRoleEngineer {
				user.RoleID = req.RoleID
			}
		}
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return model.UserResponse{}, err
	}

	return model.UserResponse{
		ID:    user.ID,
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (s *superadminService) TogglePlatformAccountStatus(ctx context.Context, id uint, isActive bool) error {
	user, err := s.userRepo.FindByID(ctx, id, []string{})
	if err != nil {
		return errors.New("account not found")
	}

	if id == 1 {
		return errors.New("cannot suspend the primary root admin")
	}

	user.IsActive = isActive
	return s.userRepo.Update(ctx, user)
}

func (s *superadminService) ListSystemRoles(ctx context.Context) ([]model.Role, error) {
	return s.roleRepo.FindSystemRoles(ctx)
}

func (s *superadminService) ListAllPermissions(ctx context.Context) ([]model.Permission, error) {
	return s.permissionRepo.FindAll(ctx)
}

func (s *superadminService) CreateSystemRole(ctx context.Context, req modelDto.CreateSystemRoleRequest, performerID uint) (model.Role, error) {
	role := &model.Role{
		TenantID:    nil,
		Name:        req.Name,
		Description: req.Description,
		BaseRole:    model.BaseRole(req.BaseRole),
		IsSystem:    true,
	}

	if role.BaseRole == "" {
		role.BaseRole = model.BaseRoleEmployee
	}

	if err := s.roleRepo.Create(ctx, role); err != nil {
		return model.Role{}, err
	}

	if len(req.PermissionIDs) > 0 {
		if err := s.roleRepo.UpdatePermissions(ctx, role.ID, req.PermissionIDs); err != nil {
			return model.Role{}, err
		}
	}

	// Audit Log
	_ = s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: performerID,
		Title:  "System Role Created",
		Action: fmt.Sprintf("Created system role: %s", role.Name),
		Status: "success",
	})

	// Fetch with permissions
	return *role, nil
}

func (s *superadminService) UpdateSystemRole(ctx context.Context, id uint, req modelDto.CreateSystemRoleRequest, performerID uint) (model.Role, error) {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return model.Role{}, errors.New("role not found")
	}

	if role.TenantID != nil {
		return model.Role{}, errors.New("cannot update tenant role via system role API")
	}

	if role.IsImmutable {
		return model.Role{}, errors.New("this system role is immutable and cannot be modified")
	}

	role.Name = req.Name
	role.Description = req.Description
	if req.BaseRole != "" {
		role.BaseRole = model.BaseRole(req.BaseRole)
	}

	if err := s.roleRepo.Update(ctx, role); err != nil {
		return model.Role{}, err
	}

	if err := s.roleRepo.UpdatePermissions(ctx, role.ID, req.PermissionIDs); err != nil {
		return model.Role{}, err
	}

	// Audit Log
	_ = s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: performerID,
		Title:  "System Role Updated",
		Action: fmt.Sprintf("Updated system role: %s", role.Name),
		Status: "success",
	})

	return *role, nil
}

func (s *superadminService) DeleteSystemRole(ctx context.Context, id uint, performerID uint) error {
	role, err := s.roleRepo.FindByID(ctx, id)
	if err != nil {
		return errors.New("role not found")
	}

	if role.TenantID != nil {
		return errors.New("cannot delete tenant role via system role API")
	}

	if role.IsImmutable {
		return errors.New("cannot delete immutable system role")
	}

	inUse, err := s.roleRepo.CheckRoleInUse(ctx, id)
	if err != nil {
		return err
	}
	if inUse {
		return errors.New("cannot delete role that is currently in use by users")
	}

	if err := s.roleRepo.Delete(ctx, id); err != nil {
		return err
	}

	// Audit Log
	_ = s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: performerID,
		Title:  "System Role Deleted",
		Action: fmt.Sprintf("Deleted system role: %s", role.Name),
		Status: "success",
	})

	return nil
}
