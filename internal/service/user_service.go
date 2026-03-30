package service

import (
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
)

type UserService interface {
	GetAllUsers(filter model.UserFilter) ([]model.User, error)
}

type userService struct {
	repo repository.UserRepository
}

func NewUserService(repo repository.UserRepository) UserService {
	return &userService{
		repo: repo,
	}
}

func (s *userService) GetAllUsers(filter model.UserFilter) ([]model.User, error) {
	// Aturan Bisnis: Jika klien tidak mengirim parameter sorting,
	// berikan nilai default agar data selalu rapi.
	if filter.OrderBy == "" {
		filter.OrderBy = "created_at" // Default urutkan berdasarkan waktu pendaftaran
	}

	if filter.Sort == "" {
		filter.Sort = "desc" // Default urutan menurun (terbaru di atas)
	}

	users, err := s.repo.FindAll(filter)
	if err != nil {
		return nil, err
	}

	return users, nil
}
