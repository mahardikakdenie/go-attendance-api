package service

import (
	"context"
	"errors"
	"time"
	dto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
)

type SubscriptionService interface {
	GetSubscriptions(ctx context.Context, page, limit int, status, search string) (dto.SubscriptionsDataResponse, error)
	GetMySubscription(ctx context.Context, tenantID uint) (*model.Subscription, error)
	UpgradeSubscription(ctx context.Context, tenantID uint, plan string) error
	RemindTenant(ctx context.Context, id uint) error
	SuspendTenant(ctx context.Context, id uint, reason string) error
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

		items = append(items, dto.SubscriptionItem{
			ID:              sub.ID,
			TenantID:        sub.TenantID,
			TenantName:      tenantName,
			TenantCode:      tenantCode,
			TenantLogo:      tenantLogo,
			Plan:            sub.Plan,
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

func (s *subscriptionService) UpgradeSubscription(ctx context.Context, tenantID uint, plan string) error {
	sub, err := s.repo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return errors.New("subscription not found")
	}

	if sub.Plan == plan {
		return errors.New("already on this plan")
	}

	var amount float64
	switch plan {
	case "Pro":
		amount = 500000
	case "Enterprise":
		amount = 1500000
	case "Basic":
		amount = 0
	default:
		return errors.New("invalid plan")
	}

	sub.Plan = plan
	sub.Amount = amount
	sub.Status = model.SubscriptionStatusActive
	sub.NextBillingDate = time.Now().AddDate(0, 1, 0)

	if err := s.repo.Update(ctx, sub); err != nil {
		return err
	}

	if sub.Tenant != nil {
		tenant := sub.Tenant
		tenant.Plan = plan
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
