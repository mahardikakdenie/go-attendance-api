package service

import (
	"context"
	"errors"
	"fmt"

	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/utils"
)

type TenantService interface {
	CreateTenant(ctx context.Context, req model.Tenant) (model.Tenant, error)
	GetAllTenants(ctx context.Context) ([]model.Tenant, error)
	GetTenantByID(ctx context.Context, id uint) (*model.Tenant, error)
	UpdateTenant(ctx context.Context, id uint, req modelDto.UpdateTenantRequest) (model.Tenant, error)
}

type tenantService struct {
	repo             repository.TenantRepository
	subscriptionRepo repository.SubscriptionRepository
	userRepo         repository.UserRepository
	notifService     NotificationService
}

func NewTenantService(
	repo repository.TenantRepository,
	subscriptionRepo repository.SubscriptionRepository,
	userRepo repository.UserRepository,
	notifService NotificationService,
) TenantService {
	return &tenantService{
		repo:             repo,
		subscriptionRepo: subscriptionRepo,
		userRepo:         userRepo,
		notifService:     notifService,
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

func (s *tenantService) UpdateTenant(ctx context.Context, id uint, req modelDto.UpdateTenantRequest) (model.Tenant, error) {
	tenant, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return model.Tenant{}, errors.New("tenant not found")
	}

	if req.Name != "" {
		tenant.Name = req.Name
	}

	if req.IsSuspended != nil {
		tenant.IsSuspended = *req.IsSuspended
	}

	if req.SuspendedReason != "" {
		tenant.SuspendedReason = req.SuspendedReason
	}

	// 1. Update Tenant Metadata
	if err := s.repo.Update(ctx, tenant); err != nil {
		return model.Tenant{}, err
	}

	// 2. Handle Subscription/Plan Change (Sync)
	if req.PlanID != 0 || req.PlanName != "" {
		var targetPlan *model.SubscriptionPlan
		var err error

		if req.PlanID != 0 {
			targetPlan, err = s.subscriptionRepo.FindPlanByID(ctx, req.PlanID)
			if err != nil {
				return model.Tenant{}, fmt.Errorf("plan with ID %d not found", req.PlanID)
			}
		} else if req.PlanName != "" {
			targetPlan, err = s.subscriptionRepo.FindPlanByName(ctx, req.PlanName)
			if err != nil {
				return model.Tenant{}, fmt.Errorf("plan with name '%s' not found", req.PlanName)
			}
		}

		if targetPlan != nil {
			sub, err := s.subscriptionRepo.FindByTenantID(ctx, id)
			if err != nil || sub == nil {
				// REGISTER NEW: No subscription record yet, create one
				sub = &model.Subscription{
					TenantID:        id,
					PlanID:          targetPlan.ID,
					Status:          model.SubscriptionStatusActive,
					BillingCycle:    model.BillingCycleMonthly,
					NextBillingDate: utils.Now().AddDate(0, 1, 0),
					Amount:          getAmountForPlan(targetPlan.Name),
				}
				if err := s.subscriptionRepo.Create(ctx, sub); err != nil {
					return model.Tenant{}, fmt.Errorf("failed to create subscription: %w", err)
				}
			} else {
				// UPDATE EXISTING: Sync plan, amount, and FORCE STATUS to Active
				sub.PlanID = targetPlan.ID
				sub.Plan = targetPlan // 🆕 Force GORM to recognize the new association
				sub.Amount = getAmountForPlan(targetPlan.Name)
				sub.Status = model.SubscriptionStatusActive // 🆕 Ensure it becomes Active (not Trial)
				if err := s.subscriptionRepo.Update(ctx, sub); err != nil {
					return model.Tenant{}, fmt.Errorf("failed to update subscription: %w", err)
				}
			}

			// Sync association back to tenant object for response
			tenant.Subscription = sub
			if tenant.Subscription != nil {
				tenant.Subscription.Plan = targetPlan
			}

			// 🆕 NOTIFICATION: Notify admins of plan change
			notifyAdmins(s.userRepo, s.notifService, ctx, id, "Subscription Plan Updated", fmt.Sprintf("Your organization plan has been updated to %s.", targetPlan.Name), model.NotificationTypeSubscription)
		}
	}

	// 3. Notification for suspension
	if req.IsSuspended != nil {
		msg := "Your organization account status has been restored to Active."
		if *req.IsSuspended {
			reason := req.SuspendedReason
			if reason == "" {
				reason = "Violation of terms or unpaid balance."
			}
			msg = fmt.Sprintf("Your organization account has been SUSPENDED. Reason: %s", reason)
		}
		notifyAdmins(s.userRepo, s.notifService, ctx, id, "Account Status Updated", msg, model.NotificationTypeSystem)
	}

	return *tenant, nil
}

// Helper: Notify admins (moved to helper for cleaner logic)
func notifyAdmins(userRepo repository.UserRepository, notifService NotificationService, ctx context.Context, tenantID uint, title, message string, category model.NotificationType) {
	admins, _, _ := userRepo.FindAll(ctx, model.UserFilter{TenantID: tenantID}, nil)
	for _, admin := range admins {
		if admin.Role != nil && (admin.Role.Name == "admin" || admin.Role.Name == "hr") {
			notifService.SendNotification(ctx, tenantID, admin.ID, title, message, category)
		}
	}
}

// Helper to get amounts from seeder-aligned names
func getAmountForPlan(name string) float64 {
	switch name {
	case "Starter":
		return 100000
	case "Business":
		return 500000
	case "Enterprise":
		return 1500000
	case "Trial":
		return 0
	default:
		return 0
	}
}
