package service

import (
	"context"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
)

type AllowancePresetService interface {
	GetAllPresets(ctx context.Context) ([]model.AllowancePreset, error)
	CreatePreset(ctx context.Context, req model.AllowancePreset) (model.AllowancePreset, error)
}

type allowancePresetService struct {
	repo repository.AllowancePresetRepository
}

func NewAllowancePresetService(repo repository.AllowancePresetRepository) AllowancePresetService {
	return &allowancePresetService{repo: repo}
}

func (s *allowancePresetService) GetAllPresets(ctx context.Context) ([]model.AllowancePreset, error) {
	return s.repo.FindAll(ctx)
}

func (s *allowancePresetService) CreatePreset(ctx context.Context, req model.AllowancePreset) (model.AllowancePreset, error) {
	err := s.repo.Create(ctx, &req)
	return req, err
}
