package service

import (
	"context"
	"go-attendance-api/internal/dto"
	"go-attendance-api/internal/repository"
)

type SuperadminService interface {
	GetOwnersWithStats(ctx context.Context, limit, offset int) ([]modelDto.OwnerWithStatsResponse, int64, error)
}

type superadminService struct {
	repo repository.SuperadminRepository
}

func NewSuperadminService(repo repository.SuperadminRepository) SuperadminService {
	return &superadminService{repo: repo}
}

func (s *superadminService) GetOwnersWithStats(ctx context.Context, limit, offset int) ([]modelDto.OwnerWithStatsResponse, int64, error) {
	if limit <= 0 {
		limit = 10
	}
	return s.repo.GetOwnersWithStats(ctx, limit, offset)
}
