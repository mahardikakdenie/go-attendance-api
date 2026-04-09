package service

import (
	"context"
	"errors"
	"fmt"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetAllUsers(ctx context.Context, filter model.UserFilter, includes []string) ([]model.UserResponse, int64, error)
	GetByID(ctx context.Context, id uint, includes []string) (model.UserResponse, error)
	GetMe(ctx context.Context, userID uint, includes []string) (model.UserResponse, error)
	GetRecentActivities(ctx context.Context, userID uint) ([]model.RecentActivityResponse, error)
	UpdateProfilePhoto(userID uint, mediaURL string) error
	CreateUser(ctx context.Context, adminID uint, req model.CreateUserRequest) (model.UserResponse, error)
}

type userService struct {
	repo         repository.UserRepository
	roleRepo     repository.RoleRepository
	activityRepo repository.RecentActivityRepository
}

func NewUserService(
	repo repository.UserRepository,
	roleRepo repository.RoleRepository,
	activityRepo repository.RecentActivityRepository,
) UserService {
	return &userService{
		repo:         repo,
		roleRepo:     roleRepo,
		activityRepo: activityRepo,
	}
}

var allowedIncludes = map[string]bool{
	"tenant":                 true,
	"tenant.tenant_settings": true,
	"attendances":            true,
	"attendances.user":       true,
	"role":                   true,
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

	users, total, err := s.repo.FindAll(ctx, filter, includes)
	if err != nil {
		return nil, 0, err
	}

	var responses []model.UserResponse
	for _, user := range users {
		responses = append(responses, mapToUserResponse(&user, includes))
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

	return mapToUserResponse(user, includes), nil
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

	return mapToUserResponse(user, includes), nil
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

func mapToUserResponse(user *model.User, includes []string) model.UserResponse {
	res := model.UserResponse{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		TenantID:    user.TenantID,
		EmployeeID:  user.EmployeeID,
		Department:  user.Department,
		MediaUrl:    user.MediaUrl,
		Address:     user.Address,
		PhoneNumber: user.PhoneNumber,
		CreatedAt:   user.CreatedAt,
	}

	if user.Role != nil {
		res.Role = &model.RoleResponse{
			ID:   user.Role.ID,
			Name: user.Role.Name,
		}
	}

	if hasInclude(includes, "tenant") && user.Tenant != nil {
		res.Tenant = &model.TenantResponse{
			ID:             user.Tenant.ID,
			Name:           user.Tenant.Name,
			TenantSettings: user.Tenant.TenantSettings,
		}
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
	admin, err := s.repo.FindByID(ctx, adminID, []string{"role"})
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

	switch adminRole {
	case "superadmin":
		tenantID = req.TenantID
		if tenantID == 0 {
			tenantID = admin.TenantID
		}
	case "admin":
		if targetRoleName != "hr" && targetRoleName != "employee" {
			return model.UserResponse{}, errors.New("admin can only create HR or Employee accounts")
		}
		tenantID = admin.TenantID
	default:
		return model.UserResponse{}, errors.New("you do not have permission to create users")
	}

	// 3. Prepare User Data
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)

	count, _ := s.repo.CountByTenantID(ctx, tenantID)
	prefix := "EMP"
	switch targetRoleName {
	case "hr":
		prefix = "HR"
	case "admin":
		prefix = "ADM"
	case "superadmin":
		prefix = "SA"
	}
	employeeID := fmt.Sprintf("%s-%03d", prefix, count+1)

	user := &model.User{
		Name:        req.Name,
		Email:       req.Email,
		Password:    string(hashedPassword),
		RoleID:      req.RoleID,
		TenantID:    tenantID,
		EmployeeID:  employeeID,
		Department:  req.Department,
		Address:     req.Address,
		PhoneNumber: req.PhoneNumber,
	}

	// 4. Use Transaction for ACID compliance
	err = s.repo.Transaction(ctx, func(txRepo repository.UserRepository) error {
		// Create User
		if err := txRepo.Create(ctx, user); err != nil {
			return err
		}

		// Log Activity (multi-table operation)
		activity := model.RecentActivity{
			UserID: adminID, // Log who created the user
			Title:  "User Management",
			Action: fmt.Sprintf("Created new user: %s (%s)", user.Name, user.EmployeeID),
			Status: "success",
		}
		
		// Note: Ideally activityRepo should also support transaction or use a shared DB instance.
		// For now, we focus on the user creation integrity.
		return s.activityRepo.Create(ctx, activity)
	})

	if err != nil {
		return model.UserResponse{}, err
	}

	// Fetch with role for response
	createdUser, _ := s.repo.FindByID(ctx, user.ID, []string{"role"})
	return mapToUserResponse(createdUser, []string{"role"}), nil
}
