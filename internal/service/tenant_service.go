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
	UpdateTenant(ctx context.Context, id uint, req model.Tenant) (model.Tenant, error)
}

type tenantService struct {
	repo             repository.TenantRepository
	subscriptionRepo repository.SubscriptionRepository
}

func NewTenantService(repo repository.TenantRepository, subscriptionRepo repository.SubscriptionRepository) TenantService {
	return &tenantService{
		repo:             repo,
		subscriptionRepo: subscriptionRepo,
	}
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

func (s *tenantService) UpdateTenant(ctx context.Context, id uint, req model.Tenant) (model.Tenant, error) {
	tenant, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return model.Tenant{}, errors.New("tenant not found")
	}

	planChanged := false
	if req.Name != "" {
		tenant.Name = req.Name
	}
	if req.Plan != "" && tenant.Plan != req.Plan {
		tenant.Plan = req.Plan
		planChanged = true
	}

	tenant.IsSuspended = req.IsSuspended
	tenant.SuspendedReason = req.SuspendedReason

	err = s.repo.Update(ctx, tenant)
	if err != nil {
		return model.Tenant{}, err
	}

	// 🔄 SYNC: If plan changed via Superadmin Manage Tenant, update the subscription record as well
	if planChanged {
		sub, err := s.subscriptionRepo.FindByTenantID(ctx, id)
		if err == nil && sub != nil {
			sub.Plan = tenant.Plan
			// Sync amount based on plan
			switch sub.Plan {
			case "Pro":
				sub.Amount = 500000
			case "Enterprise":
				sub.Amount = 1500000
			default:
				sub.Amount = 0
			}
			_ = s.subscriptionRepo.Update(ctx, sub)
		}
	}

	return *tenant, nil
}
