package service

import (
	"context"
	"errors"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
)

type UserService interface {
	GetAllUsers(ctx context.Context, filter model.UserFilter) ([]model.UserResponse, int64, error)
	GetByID(ctx context.Context, id uint) (model.UserResponse, error)
	GetMe(ctx context.Context, userID uint) (model.UserResponse, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

// ======================
// GET ALL USERS (UPDATED)
// ======================
func (s *userService) GetAllUsers(
	ctx context.Context,
	filter model.UserFilter,
) ([]model.UserResponse, int64, error) {

	// Default sorting
	if filter.OrderBy == "" {
		filter.OrderBy = "created_at"
	}

	if filter.Sort == "" {
		filter.Sort = "desc"
	}

	// Default pagination
	if filter.Limit == 0 {
		filter.Limit = 10
	}

	users, total, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	var responses []model.UserResponse
	for _, user := range users {
		responses = append(responses, mapToUserResponse(&user))
	}

	return responses, total, nil
}

// ======================
// GET BY ID (NEW 🔥)
// ======================
func (s *userService) GetByID(ctx context.Context, id uint) (model.UserResponse, error) {
	if id == 0 {
		return model.UserResponse{}, errors.New("invalid user id")
	}

	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return model.UserResponse{}, errors.New("user not found")
	}

	return mapToUserResponse(user), nil
}

// ======================
// MAPPER
// ======================
func mapToUserResponse(user *model.User) model.UserResponse {
	return model.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		TenantID:  user.TenantID,
		CreatedAt: user.CreatedAt,
	}
}

func (s *userService) GetMe(ctx context.Context, userID uint) (model.UserResponse, error) {
	user, err := s.repo.GetMe(ctx, userID)
	if err != nil {
		return model.UserResponse{}, err
	}

	return model.UserResponse{
		ID:       user.ID,
		Name:     user.Name,
		Email:    user.Email,
		Role:     user.Role,
		TenantID: user.TenantID,
	}, nil
}
