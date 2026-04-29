package service

import (
	"context"
	"errors"
	"fmt"
	dto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"time"
)

type SubscriptionService interface {
	GetSubscriptions(ctx context.Context, page, limit int, status, search string) (dto.SubscriptionsDataResponse, error)
	GetMySubscription(ctx context.Context, tenantID uint) (*model.Subscription, error)
	UpgradeSubscription(ctx context.Context, tenantID uint, planName string) error
	RemindTenant(ctx context.Context, id uint) error
	SuspendTenant(ctx context.Context, id uint, reason string) error

	// Superadmin Plan Management
	GetAllPlans(ctx context.Context) ([]model.SubscriptionPlan, error)
	GetPlanByID(ctx context.Context, id uint) (*model.SubscriptionPlan, error)
	CreatePlan(ctx context.Context, req dto.CreatePlanRequest) (*model.SubscriptionPlan, error)
	UpdatePlan(ctx context.Context, id uint, req dto.UpdatePlanRequest) (*model.SubscriptionPlan, error)
	DeletePlan(ctx context.Context, id uint) error

	// Superadmin Subscription Management
	UpdateTenantSubscription(ctx context.Context, subID uint, req dto.UpdateTenantSubscriptionRequest) (*model.Subscription, error)
}

type subscriptionService struct {
	repo       repository.SubscriptionRepository
	tenantRepo repository.TenantRepository
}

func NewSubscriptionService(repo repository.SubscriptionRepository, tenantRepo repository.TenantRepository) SubscriptionService {
	return &subscriptionService{repo: repo, tenantRepo: tenantRepo}
}

func (s *subscriptionService) GetSubscriptions(ctx context.Context, page, limit int, status, search string) (dto.SubscriptionsDataResponse, error) {
	subs, total, err := s.repo.FindAll(ctx, page, limit, status, search)
	if err != nil {
		return dto.SubscriptionsDataResponse{}, err
	}

	mrr, activeTenants, pastDueAmount, err := s.repo.GetStats(ctx)
	if err != nil {
		return dto.SubscriptionsDataResponse{}, err
	}

	var items []dto.SubscriptionItem
	for _, sub := range subs {
		activeEmployees, _ := s.repo.CountEmployees(ctx, sub.TenantID)

		var tenantName, tenantCode, tenantLogo string
		if sub.Tenant != nil {
			tenantName = sub.Tenant.Name
			tenantCode = sub.Tenant.Code
			if sub.Tenant.TenantSettings != nil {
				tenantLogo = sub.Tenant.TenantSettings.TenantLogo
			}
		}

		planName := ""
		if sub.Plan != nil {
			planName = sub.Plan.Name
		}

		items = append(items, dto.SubscriptionItem{
			ID:              sub.ID,
			TenantID:        sub.TenantID,
			TenantName:      tenantName,
			TenantCode:      tenantCode,
			TenantLogo:      tenantLogo,
			Plan:            planName,
			BillingCycle:    sub.BillingCycle,
			Amount:          sub.Amount,
			Status:          sub.Status,
			NextBillingDate: sub.NextBillingDate,
			ActiveEmployees: activeEmployees,
			CreatedAt:       sub.CreatedAt,
		})
	}

	return dto.SubscriptionsDataResponse{
		Stats: dto.SubscriptionStats{
			MRR:                 mrr,
			MRRGrowth:           "+12.5%",
			ActiveTenants:       activeTenants,
			ActiveTenantsGrowth: "+5",
			PastDueAmount:       pastDueAmount,
			PastDueGrowth:       "-2.1%",
		},
		Items: items,
		Total: total,
	}, nil
}

func (s *subscriptionService) GetMySubscription(ctx context.Context, tenantID uint) (*model.Subscription, error) {
	return s.repo.FindByTenantID(ctx, tenantID)
}

func (s *subscriptionService) UpgradeSubscription(ctx context.Context, tenantID uint, planName string) error {
	sub, err := s.repo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return errors.New("subscription not found")
	}

	if sub.Plan != nil && sub.Plan.Name == planName {
		return errors.New("already on this plan")
	}

	newPlan, err := s.repo.FindPlanByName(ctx, planName)
	if err != nil {
		return fmt.Errorf("plan %s not found", planName)
	}

	var amount float64
	switch planName {
	case "Pro":
		amount = 500000
	case "Enterprise":
		amount = 1500000
	case "Starter":
		amount = 100000
	case "Basic":
		amount = 0
	case "Trial":
		amount = 0
	default:
		return errors.New("invalid plan")
	}

	sub.PlanID = newPlan.ID
	sub.Amount = amount
	sub.Status = model.SubscriptionStatusActive
	sub.NextBillingDate = time.Now().AddDate(0, 1, 0)

	if err := s.repo.Update(ctx, sub); err != nil {
		return err
	}

	if sub.Tenant != nil {
		tenant := sub.Tenant
		tenant.Plan = planName
		return s.tenantRepo.Update(ctx, tenant)
	}

	return nil
}

func (s *subscriptionService) RemindTenant(ctx context.Context, id uint) error {
	sub, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}
	_ = sub
	return nil
}

func (s *subscriptionService) SuspendTenant(ctx context.Context, id uint, reason string) error {
	sub, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if sub.Tenant == nil {
		return errors.New("tenant not found for subscription")
	}

	sub.Status = model.SubscriptionStatusCanceled
	if err := s.repo.Update(ctx, sub); err != nil {
		return err
	}

	tenant := sub.Tenant
	tenant.IsSuspended = true
	tenant.SuspendedReason = reason

	return s.tenantRepo.Update(ctx, tenant)
}

func (s *subscriptionService) GetAllPlans(ctx context.Context) ([]model.SubscriptionPlan, error) {
	return s.repo.FindAllPlans(ctx)
}

func (s *subscriptionService) GetPlanByID(ctx context.Context, id uint) (*model.SubscriptionPlan, error) {
	return s.repo.FindPlanByID(ctx, id)
}

func (s *subscriptionService) CreatePlan(ctx context.Context, req dto.CreatePlanRequest) (*model.SubscriptionPlan, error) {
	plan := &model.SubscriptionPlan{
		Name:         req.Name,
		MaxEmployees: req.MaxEmployees,
		Features:     req.Features,
		IsActive:     true,
	}

	if err := s.repo.CreatePlan(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *subscriptionService) UpdatePlan(ctx context.Context, id uint, req dto.UpdatePlanRequest) (*model.SubscriptionPlan, error) {
	plan, err := s.repo.FindPlanByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != "" {
		plan.Name = req.Name
	}
	if req.MaxEmployees >= 0 {
		plan.MaxEmployees = req.MaxEmployees
	}
	if req.Features != nil {
		plan.Features = req.Features
	}
	if req.IsActive != nil {
		plan.IsActive = *req.IsActive
	}

	if err := s.repo.UpdatePlan(ctx, plan); err != nil {
		return nil, err
	}
	return plan, nil
}

func (s *subscriptionService) DeletePlan(ctx context.Context, id uint) error {
	return s.repo.DeletePlan(ctx, id)
}

func (s *subscriptionService) UpdateTenantSubscription(ctx context.Context, subID uint, req dto.UpdateTenantSubscriptionRequest) (*model.Subscription, error) {
	sub, err := s.repo.FindByID(ctx, subID)
	if err != nil {
		return nil, err
	}

	if req.PlanID != 0 {
		plan, err := s.repo.FindPlanByID(ctx, req.PlanID)
		if err != nil {
			return nil, fmt.Errorf("invalid plan id: %v", err)
		}
		sub.PlanID = plan.ID
		// Sync Tenant string plan for UI compatibility
		if sub.Tenant != nil {
			sub.Tenant.Plan = plan.Name
			_ = s.tenantRepo.Update(ctx, sub.Tenant)
		}
	}

	if req.Status != "" {
		sub.Status = model.SubscriptionStatus(req.Status)
	}

	if req.Amount >= 0 {
		sub.Amount = req.Amount
	}

	if !req.NextBillingDate.IsZero() {
		sub.NextBillingDate = req.NextBillingDate
	}

	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, err
	}
	return sub, nil
}
