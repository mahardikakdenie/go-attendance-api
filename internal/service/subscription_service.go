package service

import (
	"context"
	"errors"
	"fmt"
	dto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/utils"
)

type SubscriptionService interface {
	GetSubscriptions(ctx context.Context, page, limit int, status, search string) (dto.SubscriptionsDataResponse, error)
	GetMySubscription(ctx context.Context, tenantID uint) (*model.Subscription, error)
	UpgradeSubscription(ctx context.Context, tenantID uint, req dto.UpgradeRequest) error
	RemindTenant(ctx context.Context, id uint) error
	SuspendTenant(ctx context.Context, id uint, reason string) error
	ReactivateSubscription(ctx context.Context, subID uint, superadminID uint, ip string) (*model.Subscription, error)

	// Superadmin Plan Management
	GetAllPlans(ctx context.Context) ([]model.SubscriptionPlan, error)
	GetPlanByID(ctx context.Context, id uint) (*model.SubscriptionPlan, error)
	CreatePlan(ctx context.Context, req dto.CreatePlanRequest) (*model.SubscriptionPlan, error)
	UpdatePlan(ctx context.Context, id uint, req dto.UpdatePlanRequest) (*model.SubscriptionPlan, error)
	DeletePlan(ctx context.Context, id uint) error

	// Superadmin Subscription Management
	UpdateTenantSubscription(ctx context.Context, subID uint, req dto.UpdateTenantSubscriptionRequest) (*model.Subscription, error)

	GetAllFeatures(ctx context.Context) ([]model.SubscriptionFeature, error)
}

type subscriptionService struct {
	repo         repository.SubscriptionRepository
	tenantRepo   repository.TenantRepository
	userRepo     repository.UserRepository
	auditRepo    repository.AuditLogRepository
	notifService NotificationService
}

func NewSubscriptionService(
	repo repository.SubscriptionRepository,
	tenantRepo repository.TenantRepository,
	userRepo repository.UserRepository,
	auditRepo repository.AuditLogRepository,
	notifService NotificationService,
) SubscriptionService {
	return &subscriptionService{
		repo:         repo,
		tenantRepo:   tenantRepo,
		userRepo:     userRepo,
		auditRepo:    auditRepo,
		notifService: notifService,
	}
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
		maxEmployees := int64(0)
		if sub.Plan != nil {
			planName = sub.Plan.Name
			maxEmployees = int64(sub.Plan.MaxEmployees)
		}

		remainingLimit := maxEmployees - activeEmployees
		if remainingLimit < 0 {
			remainingLimit = 0
		}

		items = append(items, dto.SubscriptionItem{
			ID:                      sub.ID,
			TenantID:                sub.TenantID,
			TenantName:              tenantName,
			TenantCode:              tenantCode,
			TenantLogo:              tenantLogo,
			Plan:                    planName,
			BillingCycle:            sub.BillingCycle,
			Amount:                  sub.Amount,
			Status:                  sub.Status,
			NextBillingDate:         sub.NextBillingDate,
			ActiveEmployees:         activeEmployees,
			RemainingEmployeesLimit: remainingLimit,
			CreatedAt:               sub.CreatedAt,
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
	sub, err := s.repo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 🆕 Dynamic Status Check: If NextBillingDate has passed, set to Non-Active
	if sub.Status != model.SubscriptionStatusNonActive && sub.Status != model.SubscriptionStatusCanceled {
		if utils.Now().After(sub.NextBillingDate) {
			sub.Status = model.SubscriptionStatusNonActive
			_ = s.repo.Update(ctx, sub)
		}
	}

	return sub, nil
}

func (s *subscriptionService) UpgradeSubscription(ctx context.Context, tenantID uint, req dto.UpgradeRequest) error {
	sub, err := s.repo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return errors.New("subscription not found")
	}

	var newPlan *model.SubscriptionPlan
	if req.PlanID != 0 {
		newPlan, err = s.repo.FindPlanByID(ctx, req.PlanID)
		if err != nil {
			return errors.New("plan not found")
		}
	} else if req.Plan != "" {
		newPlan, err = s.repo.FindPlanByName(ctx, req.Plan)
		if err != nil {
			return fmt.Errorf("plan %s not found", req.Plan)
		}
	} else {
		return errors.New("plan or plan_id is required")
	}

	if sub.PlanID == newPlan.ID {
		return errors.New("already on this plan")
	}

	sub.PlanID = newPlan.ID
	sub.Amount = newPlan.Price
	sub.Status = model.SubscriptionStatusActive

	duration := 30 // Default 30 days if not set
	if newPlan.Days > 0 {
		duration = newPlan.Days
	}
	sub.NextBillingDate = utils.Now().AddDate(0, 0, duration)

	if err := s.repo.Update(ctx, sub); err != nil {
		return err
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

func (s *subscriptionService) ReactivateSubscription(ctx context.Context, subID uint, superadminID uint, ip string) (*model.Subscription, error) {
	// 1. Validate ID exists
	sub, err := s.repo.FindByID(ctx, subID)
	if err != nil {
		return nil, errors.New("subscription not found")
	}

	// 2. Validate current status is Canceled
	if sub.Status != model.SubscriptionStatusCanceled {
		return nil, fmt.Errorf("only canceled subscriptions can be reactivated (current: %s)", sub.Status)
	}

	oldStatus := string(sub.Status)

	// 3. Change status to Active
	sub.Status = model.SubscriptionStatusActive

	// 4. Ensure tenant is unsuspended
	if sub.Tenant != nil {
		tenant := sub.Tenant
		tenant.IsSuspended = false
		tenant.SuspendedReason = ""
		_ = s.tenantRepo.Update(ctx, tenant)
	}

	// 5. Recalculate billing date if passed
	now := utils.Now()
	if sub.NextBillingDate.Before(now) {
		duration := 30 // Default 30 days
		if sub.Plan != nil && sub.Plan.Days > 0 {
			duration = sub.Plan.Days
		}
		sub.NextBillingDate = now.AddDate(0, 0, duration)
	}

	// Save updates
	if err := s.repo.Update(ctx, sub); err != nil {
		return nil, fmt.Errorf("failed to update subscription: %w", err)
	}

	// 6. Notify Tenant Owner
	admins, _, _ := s.userRepo.FindAll(ctx, model.UserFilter{TenantID: sub.TenantID}, nil)
	for _, admin := range admins {
		if admin.Role != nil && (admin.Role.Name == "admin" || admin.Role.Name == "hr") {
			s.notifService.SendNotification(ctx, sub.TenantID, admin.ID, "Subscription Restored", "Your organization subscription has been reactivated by System Administrator. Full access restored.", model.NotificationTypeSubscription)
		}
	}

	// 7. Log to Audit Logs
	_ = s.auditRepo.Create(ctx, &model.AuditLog{
		UserID:    superadminID,
		Action:    "SUBSCRIPTION_REACTIVATED",
		Entity:    "subscription",
		EntityID:  fmt.Sprintf("%d", sub.ID),
		OldValue:  oldStatus,
		NewValue:  string(sub.Status),
		IPAddress: ip,
	})

	return sub, nil
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
		Price:        req.Price,
		Days:         req.Days,
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
	if req.Price >= 0 {
		plan.Price = req.Price
	}
	if req.Days >= 0 {
		plan.Days = req.Days
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

func (s *subscriptionService) GetAllFeatures(ctx context.Context) ([]model.SubscriptionFeature, error) {
	return s.repo.FindAllFeatures(ctx)
}
