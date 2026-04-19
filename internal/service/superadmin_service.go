package service

import (
	"context"
	"errors"
	"fmt"
	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/utils"
	"log"
	"sort"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type SuperadminService interface {
	GetOwnersWithStats(ctx context.Context, limit, offset int) ([]modelDto.OwnerWithStatsResponse, int64, error)
	GetPlatformAccounts(ctx context.Context, search string, limit, offset int) ([]model.UserResponse, int64, error)
	CreatePlatformAccount(ctx context.Context, req model.CreateUserRequest, performerID uint) (model.UserResponse, error)
	UpdatePlatformAccount(ctx context.Context, id uint, req model.CreateUserRequest, performerID uint) (model.UserResponse, error)
	TogglePlatformAccountStatus(ctx context.Context, id uint, isActive bool) error

	// System Role Management
	ListSystemRoles(ctx context.Context) ([]model.Role, error)
	ListAllPermissions(ctx context.Context) ([]modelDto.PermissionModule, error)
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

func (s *superadminService) toUserResponse(user model.User) model.UserResponse {
	resp := model.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		TenantID:  user.TenantID,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		BaseRole:  user.Role.BaseRole,
	}
	if user.Role.ID != 0 {
		resp.Role = &model.RoleResponse{
			ID:       user.Role.ID,
			Name:     user.Role.Name,
			BaseRole: user.Role.BaseRole,
		}
	}
	return resp
}

func (s *superadminService) isSystemRole(baseRole model.BaseRole) bool {
	return baseRole == model.BaseRoleSuperAdmin || baseRole == model.BaseRoleSupport || baseRole == model.BaseRoleEngineer
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

	// ISSUE-002: Pre-allocate with capacity and ensure it's not nil
	responses := make([]model.UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, s.toUserResponse(user))
	}

	return responses, total, nil
}

func (s *superadminService) CreatePlatformAccount(ctx context.Context, req model.CreateUserRequest, performerID uint) (model.UserResponse, error) {
	// Platform accounts are always tenant 1 (HQ)
	var HQTenantID = 1
	req.TenantID = uint(HQTenantID)

	// Validate Role is a system role (SUPERADMIN, SUPPORT, ENGINEER)
	role, err := s.roleRepo.FindByID(ctx, req.RoleID)
	if err != nil || role == nil {
		return model.UserResponse{}, errors.New("invalid role")
	}

	if !s.isSystemRole(role.BaseRole) {
		return model.UserResponse{}, errors.New("role must be a system role (SUPERADMIN, SUPPORT, or ENGINEER)")
	}

	password := req.Password
	if password == "" {
		password = utils.GenerateRandomString(12)
	}

	// ISSUE-001: Validate length and handle error
	if len([]byte(password)) > 72 {
		return model.UserResponse{}, errors.New("password exceeds maximum length of 72 bytes")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return model.UserResponse{}, fmt.Errorf("failed to hash password: %w", err)
	}

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

	// Audit Log (ISSUE-008: use performerID)
	_ = s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: performerID,
		Title:  "Platform Administration",
		Action: fmt.Sprintf("Created platform account: %s (%s)", user.Name, role.BaseRole),
		Status: "success",
	})

	// Send Email (ISSUE-007: add timeout and logging)
	emailHtml := utils.GetWelcomeEmailTemplate(user.Name, user.Email, password, "Attendance System HQ", "")
	subject := "Platform Administrator Account Created"
	go func() {
		emailCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := utils.SendEmail(emailCtx, []string{user.Email}, subject, emailHtml); err != nil {
			log.Printf("warn: failed to send welcome email to %s: %v", user.Email, err)
		}
	}()

	// Ensure role is loaded for the response
	user.Role = role
	return s.toUserResponse(*user), nil
}

func (s *superadminService) UpdatePlatformAccount(ctx context.Context, id uint, req model.CreateUserRequest, performerID uint) (model.UserResponse, error) {
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
		if err != nil || role == nil {
			return model.UserResponse{}, errors.New("invalid role")
		}
		// ISSUE-003: explicitly check if system role
		if !s.isSystemRole(role.BaseRole) {
			return model.UserResponse{}, errors.New("role must be a system role")
		}
		user.RoleID = req.RoleID
		user.Role = role // update loaded role for response
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return model.UserResponse{}, err
	}

	// Audit Log
	_ = s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: performerID,
		Title:  "Platform Administration",
		Action: fmt.Sprintf("Updated platform account: %s", user.Name),
		Status: "success",
	})

	return s.toUserResponse(*user), nil
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

func (s *superadminService) ListAllPermissions(ctx context.Context) ([]modelDto.PermissionModule, error) {
	permissions, err := s.permissionRepo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// Module name mapping for better UI display
	moduleNames := map[string]string{
		"attendance":   "Attendance & Monitoring",
		"leave":        "Leave Management",
		"overtime":     "Overtime & Extra Hours",
		"payroll":      "Payroll & Finance",
		"user":         "User Management",
		"tenant":       "Organization & SaaS",
		"subscription": "Plans & Billing",
		"role":         "Roles & Permissions",
		"support":      "Support & Helpdesk",
		"project":      "Project Management",
		"timesheet":    "Time Tracking",
		"finance":      "Finance Operations",
		"performance":  "Performance & Goals",
	}

	// Group permissions by module
	modulesMap := make(map[string][]modelDto.PermissionResponse)
	for _, p := range permissions {
		resp := modelDto.PermissionResponse{
			ID:          p.ID,
			Module:      p.Module,
			Action:      p.Action,
			Description: p.Description,
		}
		modulesMap[p.Module] = append(modulesMap[p.Module], resp)
	}

	// Convert map to slice of PermissionModule
	var result []modelDto.PermissionModule
	for moduleKey, perms := range modulesMap {
		name := moduleNames[moduleKey]
		if name == "" {
			name = moduleKey // Fallback to raw key if not mapped
		}

		result = append(result, modelDto.PermissionModule{
			Name:        name,
			Key:         moduleKey,
			Permissions: perms,
		})
	}

	// ISSUE-005: Sort for deterministic response
	sort.Slice(result, func(i, j int) bool {
		return result[i].Key < result[j].Key
	})

	return result, nil
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

	// ISSUE-006: Use transactional method
	if err := s.roleRepo.CreateWithPermissions(ctx, role, req.PermissionIDs); err != nil {
		return model.Role{}, err
	}

	// Audit Log
	_ = s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: performerID,
		Title:  "System Role Created",
		Action: fmt.Sprintf("Created system role: %s", role.Name),
		Status: "success",
	})

	// Fetch with permissions
	updatedRole, _ := s.roleRepo.FindByID(ctx, role.ID)
	if updatedRole != nil {
		return *updatedRole, nil
	}
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

	// ISSUE-006: Use transactional method
	if err := s.roleRepo.UpdateWithPermissions(ctx, role, req.PermissionIDs); err != nil {
		return model.Role{}, err
	}

	// Audit Log
	_ = s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: performerID,
		Title:  "System Role Updated",
		Action: fmt.Sprintf("Updated system role: %s", role.Name),
		Status: "success",
	})

	// Fetch with permissions
	updatedRole, _ := s.roleRepo.FindByID(ctx, role.ID)
	if updatedRole != nil {
		return *updatedRole, nil
	}
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
