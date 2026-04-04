package service

import (
	"context"
	"errors"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
)

type UserService interface {
	GetAllUsers(ctx context.Context, filter model.UserFilter, includes []string) ([]model.UserResponse, int64, error)
	GetByID(ctx context.Context, id uint, includes []string) (model.UserResponse, error)
	GetMe(ctx context.Context, userID uint, includes []string) (model.UserResponse, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

var allowedIncludes = map[string]bool{
	"tenant":           true,
	"attendances":      true,
	"attendances.user": true,
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

func mapToUserResponse(user *model.User, includes []string) model.UserResponse {
	res := model.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		TenantID:  user.TenantID,
		CreatedAt: user.CreatedAt,
	}

	if hasInclude(includes, "tenant") && user.Tenant != nil {
		res.Tenant = &model.TenantResponse{
			ID:   user.Tenant.ID,
			Name: user.Tenant.Name,
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

	return res
}
