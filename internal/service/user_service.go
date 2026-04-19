package service

import (
	"context"
	"errors"
	"fmt"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/utils"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetAllUsers(ctx context.Context, filter model.UserFilter, includes []string) ([]model.UserResponse, int64, error)
	GetByID(ctx context.Context, id uint, includes []string) (model.UserResponse, error)
	GetMe(ctx context.Context, userID uint, includes []string) (model.UserResponse, error)
	GetRecentActivities(ctx context.Context, userID uint) ([]model.RecentActivityResponse, error)
	UpdateProfilePhoto(userID uint, mediaURL string) error
	CreateUser(ctx context.Context, adminID uint, req model.CreateUserRequest) (model.UserResponse, error)
	GetAllowedRoleIDs(ctx context.Context, userID uint) ([]uint, error)
}

type userService struct {
	repo          repository.UserRepository
	roleRepo      repository.RoleRepository
	activityRepo  repository.RecentActivityRepository
	hierarchyRepo repository.RoleHierarchyRepository
	hrOpsRepo     repository.HrOpsRepository
	leaveRepo     repository.LeaveRepository
	profileRepo   repository.UserPayrollProfileRepository
}

func NewUserService(
	repo repository.UserRepository,
	roleRepo repository.RoleRepository,
	activityRepo repository.RecentActivityRepository,
	hierarchyRepo repository.RoleHierarchyRepository,
	hrOpsRepo repository.HrOpsRepository,
	leaveRepo repository.LeaveRepository,
	profileRepo repository.UserPayrollProfileRepository,
) UserService {
	return &userService{
		repo:          repo,
		roleRepo:      roleRepo,
		activityRepo:  activityRepo,
		hierarchyRepo: hierarchyRepo,
		hrOpsRepo:     hrOpsRepo,
		leaveRepo:     leaveRepo,
		profileRepo:   profileRepo,
	}
}

var allowedIncludes = map[string]bool{
	"tenant":                 true,
	"tenant.tenant_settings": true,
	"tenant_setting":          true,
	"attendances":            true,
	"attendances.user":       true,
	"role":                   true,
	"role.permissions":       true,
	"recent_activities":      true,
}

func filterIncludes(includes []string) []string {
	var result []string
	for _, inc := range includes {
		if allowedIncludes[inc] {
			result = append(result, inc)
		}
	}
	return result
}

func hasInclude(includes []string, key string) bool {
	for _, inc := range includes {
		if inc == key {
			return true
		}
	}
	return false
}

func (s *userService) GetAllUsers(
	ctx context.Context,
	filter model.UserFilter,
	includes []string,
) ([]model.UserResponse, int64, error) {

	if filter.OrderBy == "" {
		filter.OrderBy = "created_at"
	}

	if filter.Sort == "" {
		filter.Sort = "desc"
	}

	if filter.Limit == 0 {
		filter.Limit = 10
	}

	includes = filterIncludes(includes)

	// Apply Hierarchical Scoping
	if filter.RequesterID != 0 {
		allowedRoleIDs, _ := s.GetAllowedRoleIDs(ctx, filter.RequesterID)
		filter.AllowedRoleIDs = allowedRoleIDs
	}

	users, total, err := s.repo.FindAll(ctx, filter, includes)
	if err != nil {
		return nil, 0, err
	}

	var responses []model.UserResponse
	for _, user := range users {
		responses = append(responses, mapToUserResponse(&user, includes, nil))
	}

	return responses, total, nil
}

func (s *userService) GetByID(
	ctx context.Context,
	id uint,
	includes []string,
) (model.UserResponse, error) {

	if id == 0 {
		return model.UserResponse{}, errors.New("invalid user id")
	}

	includes = filterIncludes(includes)

	user, err := s.repo.FindByID(ctx, id, includes)
	if err != nil {
		return model.UserResponse{}, errors.New("user not found")
	}

	return mapToUserResponse(user, includes, nil), nil
}

func (s *userService) GetMe(
	ctx context.Context,
	userID uint,
	includes []string,
) (model.UserResponse, error) {

	if userID == 0 {
		return model.UserResponse{}, errors.New("invalid user id")
	}

	includes = filterIncludes(includes)

	user, err := s.repo.GetMe(ctx, userID, includes)
	if err != nil {
		return model.UserResponse{}, err
	}

	// Resolve active shift for today
	var activeShift *model.WorkShift
	now := time.Now().In(time.FixedZone("WIB", 7*3600))

	// 0. Check Approved Leave
	isOnLeave, _ := s.leaveRepo.CheckOnLeave(ctx, userID, now)
	if isOnLeave {
		activeShift = &model.WorkShift{
			Name:      "Sedang Cuti",
			StartTime: "00:00",
			EndTime:   "23:59",
			Type:      "Morning",
			Color:     "bg-orange-500",
		}
	}

	if activeShift == nil {
		// 1. Check explicit roster for today
		rosters, _ := s.hrOpsRepo.FindRoster(ctx, user.TenantID, userID, now, now)
		if len(rosters) > 0 && rosters[0].ShiftID != nil {
			activeShift = rosters[0].Shift
		}
	}

	// 2. If no roster entry, check for default shift
	if activeShift == nil {
		activeShift, _ = s.hrOpsRepo.FindDefaultShift(ctx, user.TenantID)
	}

	// 3. If still no shift, mock one using Tenant Settings
	if activeShift == nil && user.Tenant != nil && user.Tenant.TenantSettings != nil {
		setting := user.Tenant.TenantSettings
		activeShift = &model.WorkShift{
			Name:      "Standard Shift",
			StartTime: setting.ClockInStartTime,
			EndTime:   setting.ClockOutEndTime,
			Type:      "Morning",
			Color:     "bg-blue-500",
		}
	}

	return mapToUserResponse(user, includes, activeShift), nil
}

func (s *userService) GetRecentActivities(ctx context.Context, userID uint) ([]model.RecentActivityResponse, error) {
	activities, err := s.activityRepo.FindByUserID(ctx, userID, 10)
	if err != nil {
		return nil, err
	}

	var responses []model.RecentActivityResponse
	for _, act := range activities {
		responses = append(responses, model.RecentActivityResponse{
			ID:        act.ID,
			Title:     act.Title,
			Action:    act.Action,
			Status:    act.Status,
			CreatedAt: act.CreatedAt,
		})
	}

	return responses, nil
}

func mapToUserResponse(user *model.User, includes []string, shift *model.WorkShift) model.UserResponse {
	res := model.UserResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		TenantID:    user.TenantID,
		PositionID:  user.PositionID,
		EmployeeID:  user.EmployeeID,
		Department:  user.Department,
		MediaUrl:    user.MediaUrl,
		Address:     user.Address,
		PhoneNumber: user.PhoneNumber,
		CreatedAt:   user.CreatedAt,
		ExpenseQuota: user.ExpenseQuota,
	}

	if user.Position != nil {
		res.Position = user.Position.Name
	}

	if shift != nil {
		res.Shift = &model.WorkShiftResponse{
			ID:        shift.ID,
			Name:      shift.Name,
			StartTime: shift.StartTime,
			EndTime:   shift.EndTime,
			Type:      shift.Type,
			Color:     shift.Color,
			IsDefault: shift.IsDefault,
		}
	}

	if user.Role != nil {
		res.Role = &model.RoleResponse{
			ID:          user.Role.ID,
			Name:        user.Role.Name,
			Description: user.Role.Description,
			BaseRole:    user.Role.BaseRole,
			IsSystem:    user.Role.IsSystem,
		}

		permissions := make([]string, len(user.Role.Permissions))
		for i, p := range user.Role.Permissions {
			permissions[i] = p.ID
		}
		res.Permissions = permissions
		res.BaseRole = user.Role.BaseRole
		res.IsOwner = user.Role.BaseRole == model.BaseRoleAdmin
	}

	if (hasInclude(includes, "tenant") || hasInclude(includes, "tenant_setting")) && user.Tenant != nil {
		res.Tenant = &model.TenantResponse{
			ID:             user.Tenant.ID,
			Name:           user.Tenant.Name,
			TenantSettings: user.Tenant.TenantSettings,
		}
		res.TenantSetting = user.Tenant.TenantSettings
	}

	if hasInclude(includes, "attendances") {
		for _, att := range user.Attendances {
			res.Attendances = append(res.Attendances, model.AttendanceResponse{
				ID:                att.ID,
				UserID:            att.UserID,
				ClockInTime:       att.ClockInTime,
				ClockOutTime:      att.ClockOutTime,
				ClockInLatitude:   att.ClockInLatitude,
				ClockInLongitude:  att.ClockInLongitude,
				ClockOutLatitude:  att.ClockOutLatitude,
				ClockOutLongitude: att.ClockOutLongitude,
				ClockInMediaUrl:   att.ClockInMediaUrl,
				ClockOutMediaUrl:  att.ClockOutMediaUrl,
				Status:            att.Status,
				CreatedAt:         att.ClockInTime, // Using ClockInTime as CreatedAt for consistency if needed, or actual CreatedAt if available
			})
		}
	}

	if hasInclude(includes, "recent_activities") {
		for _, act := range user.RecentActivities {
			res.RecentActivities = append(res.RecentActivities, model.RecentActivityResponse{
				ID:        act.ID,
				Title:     act.Title,
				Action:    act.Action,
				Status:    act.Status,
				CreatedAt: act.CreatedAt,
			})
		}
	}

	return res
}

func (s *userService) UpdateProfilePhoto(userID uint, mediaURL string) error {
	user, err := s.repo.FindByID(context.Background(), userID, []string{})
	if err != nil {
		return errors.New("user not found")
	}

	user.MediaUrl = mediaURL

	if err := s.repo.Update(context.Background(), user); err != nil {
		return errors.New("failed to update profile photo")
	}

	return nil
}

func (s *userService) CreateUser(ctx context.Context, adminID uint, req model.CreateUserRequest) (model.UserResponse, error) {
	// 1. Get Admin/Creator Info
	admin, err := s.repo.FindByID(ctx, adminID, []string{"role", "tenant.tenant_settings"})
	if err != nil {
		return model.UserResponse{}, errors.New("creator not found")
	}

	// 2. Validate Role Permissions
	targetRole, err := s.roleRepo.FindByID(ctx, req.RoleID)
	if err != nil || targetRole == nil {
		return model.UserResponse{}, errors.New("invalid target role")
	}

	adminRole := admin.Role.Name
	targetRoleName := targetRole.Name

	var tenantID uint
	var companyName string
	var logoURL string

	if admin.Tenant != nil {
		companyName = admin.Tenant.Name
		if admin.Tenant.TenantSettings != nil {
			logoURL = admin.Tenant.TenantSettings.TenantLogo
		}
	}

	switch adminRole {
	case "superadmin":
		// Superadmin can create any role in any tenant
		tenantID = req.TenantID
		if tenantID == 0 {
			tenantID = admin.TenantID
		}
		// If Superadmin creates for different tenant, we might need to fetch that tenant name for email
		if tenantID != admin.TenantID {
			t, _ := s.repo.FindTenantByID(ctx, tenantID)
			if t != nil {
				companyName = t.Name
				if t.TenantSettings != nil {
					logoURL = t.TenantSettings.TenantLogo
				}
			}
		}
	case "admin":
		if targetRoleName != "hr" && targetRoleName != "employee" && targetRoleName != "finance" {
			return model.UserResponse{}, errors.New("admin can only create HR, Finance, or Employee accounts")
		}
		tenantID = admin.TenantID
	default:
		return model.UserResponse{}, errors.New("you do not have permission to create users")
	}

	// 3. Prepare User Data & Password Generation
	password := req.Password
	isSystemGenerated := false
	if password == "" {
		password = utils.GenerateRandomString(8)
		isSystemGenerated = true
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	count, _ := s.repo.CountByTenantID(ctx, tenantID)
	prefix := "EMP"
	switch targetRoleName {
	case "hr":
		prefix = "HR"
	case "admin":
		prefix = "ADM"
	case "superadmin":
		prefix = "SA"
	case "finance":
		prefix = "FIN"
	}
	employeeID := fmt.Sprintf("%s-%03d", prefix, count+1)

	user := &model.User{
		Name:               req.Name,
		Email:              req.Email,
		Password:           string(hashedPassword),
		RoleID:             req.RoleID,
		TenantID:           tenantID,
		EmployeeID:         employeeID,
		Department:         req.Department,
		Address:            req.Address,
		PhoneNumber:        req.PhoneNumber,
		IsSystemCreated:    isSystemGenerated,
		MustChangePassword: isSystemGenerated,
	}

	// 4. Use Transaction for ACID compliance
	err = s.repo.Transaction(ctx, func(txRepo repository.UserRepository) error {
		if err := txRepo.Create(ctx, user); err != nil {
			return err
		}

		// 🆕 Automatic creation of Payroll Profile baseline
		profile := &model.UserPayrollProfile{
			UserID: user.ID,
		}
		// We use s.profileRepo.Upsert or similar if we have it, 
		// but since it's a new user, a simple Create or Upsert is fine.
		if err := s.profileRepo.Upsert(ctx, profile); err != nil {
			return fmt.Errorf("failed to create user payroll profile: %v", err)
		}

		activity := model.RecentActivity{
			UserID: adminID,
			Title:  "User Management",
			Action: fmt.Sprintf("Created new user: %s (%s)", user.Name, user.EmployeeID),
			Status: "success",
		}
		_ = s.activityRepo.Create(ctx, &activity)

		// 5. Send Branded Welcome Email
		emailHtml := utils.GetWelcomeEmailTemplate(user.Name, user.Email, password, companyName, logoURL)
		subject := fmt.Sprintf("Welcome to %s - Your Account Details", companyName)

		go func() {
			emailCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_ = utils.SendEmail(emailCtx, []string{user.Email}, subject, emailHtml)
		}()

		return nil
	})

	if err != nil {
		return model.UserResponse{}, err
	}

	createdUser, _ := s.repo.FindByID(ctx, user.ID, []string{"role"})
	return mapToUserResponse(createdUser, []string{"role"}, nil), nil
}

func (s *userService) GetAllowedRoleIDs(ctx context.Context, userID uint) ([]uint, error) {
	user, err := s.repo.FindByID(ctx, userID, []string{"role"})
	if err != nil {
		return nil, err
	}

	// Superadmin logic: If it's ADMIN base role, they can see EVERYTHING in their tenant
	// So we return empty list to signify "no filter" (Repo handles this)
	if user.Role != nil && user.Role.BaseRole == model.BaseRoleAdmin {
		return []uint{}, nil
	}

	// For other roles, get descendants
	return s.hierarchyRepo.GetAllDescendantRoleIDs(ctx, user.RoleID)
}
