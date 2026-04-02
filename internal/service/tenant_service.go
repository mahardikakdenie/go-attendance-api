package service

import (
	"context"
	"errors"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
)

type TenantService interface {
	CreateTenant(ctx context.Context, req model.Tenant) (model.Tenant, error)
	GetAllTenants(ctx context.Context) ([]model.Tenant, error)
	GetTenantByID(ctx context.Context, id uint) (*model.Tenant, error)
}

type tenantService struct {
	repo repository.TenantRepository
}

func NewTenantService(repo repository.TenantRepository) TenantService {
	return &tenantService{repo: repo}
}

func (s *tenantService) CreateTenant(ctx context.Context, req model.Tenant) (model.Tenant, error) {
	if req.Name == "" || req.Code == "" {
		return model.Tenant{}, errors.New("name and code are required")
	}

	err := s.repo.Create(ctx, &req)
	if err != nil {
		return model.Tenant{}, err
	}

	return req, nil
}

func (s *tenantService) GetAllTenants(ctx context.Context) ([]model.Tenant, error) {
	return s.repo.FindAll(ctx)
}

func (s *tenantService) GetTenantByID(ctx context.Context, id uint) (*model.Tenant, error) {
	if id == 0 {
		return nil, errors.New("invalid tenant id")
	}

	return s.repo.FindByID(ctx, id)
}
