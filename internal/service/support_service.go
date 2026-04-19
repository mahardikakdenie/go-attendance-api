package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/utils"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type SupportService interface {
	// Public
	CreateTrialRequest(ctx context.Context, req modelDto.CreateTrialRequestRequest) (modelDto.TrialRequestResponse, error)

	// Admin (Tenant 1)
	GetAllTrialRequests(ctx context.Context) ([]modelDto.TrialRequestResponse, error)
	UpdateTrialStatus(ctx context.Context, id uuid.UUID, status model.TrialRequestStatus) error

	GetAllSupportMessages(ctx context.Context) ([]modelDto.SupportMessageResponse, error)
	UpdateSupportStatus(ctx context.Context, id uuid.UUID, status model.SupportStatus) error

	// Superadmin Only
	GetAllProvisioningTickets(ctx context.Context) ([]modelDto.ProvisioningTicketResponse, error)
	ExecuteProvisioning(ctx context.Context, ticketID uuid.UUID, adminID uint) error

	// Tenant User
	CreateSupportMessage(ctx context.Context, tenantID uint, userID uint, req modelDto.CreateSupportMessageRequest) (modelDto.SupportMessageResponse, error)
}

type supportService struct {
	repo              repository.SupportRepository
	tenantRepo        repository.TenantRepository
	userRepo          repository.UserRepository
	roleRepo          repository.RoleRepository
	subscriptionRepo  repository.SubscriptionRepository
	tenantSettingRepo repository.TenantSettingRepository
	profileRepo       repository.UserPayrollProfileRepository
}

func NewSupportService(
	repo repository.SupportRepository,
	tenantRepo repository.TenantRepository,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	subscriptionRepo repository.SubscriptionRepository,
	tenantSettingRepo repository.TenantSettingRepository,
	profileRepo repository.UserPayrollProfileRepository,
) SupportService {
	return &supportService{
		repo:              repo,
		tenantRepo:        tenantRepo,
		userRepo:          userRepo,
		roleRepo:          roleRepo,
		subscriptionRepo:  subscriptionRepo,
		tenantSettingRepo: tenantSettingRepo,
		profileRepo:       profileRepo,
	}
}

func (s *supportService) CreateTrialRequest(ctx context.Context, req modelDto.CreateTrialRequestRequest) (modelDto.TrialRequestResponse, error) {
	trial := &model.TrialRequest{
		CompanyName:        req.CompanyName,
		ContactName:        req.ContactName,
		Email:              req.Email,
		PhoneNumber:        req.PhoneNumber,
		EmployeeCountRange: req.EmployeeCountRange,
		Industry:           req.Industry,
		Status:             model.TrialRequestStatusNew,
	}

	if err := s.repo.CreateTrialRequest(ctx, trial); err != nil {
		return modelDto.TrialRequestResponse{}, err
	}

	// Kirim email konfirmasi secara asinkron
	go func() {
		emailCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		emailHtml := utils.GetTrialConfirmationEmailTemplate(trial.ContactName, trial.CompanyName)
		subject := "Trial Request Received - Attendance System"
		_ = utils.SendEmail(emailCtx, []string{trial.Email}, subject, emailHtml)
	}()

	return mapToTrialRequestResponse(trial), nil
}

func (s *supportService) GetAllTrialRequests(ctx context.Context) ([]modelDto.TrialRequestResponse, error) {
	trials, err := s.repo.FindAllTrialRequests(ctx)
	if err != nil {
		return nil, err
	}

	var responses []modelDto.TrialRequestResponse
	for _, t := range trials {
		responses = append(responses, mapToTrialRequestResponse(&t))
	}
	return responses, nil
}

func (s *supportService) UpdateTrialStatus(ctx context.Context, id uuid.UUID, status model.TrialRequestStatus) error {
	trial, err := s.repo.FindTrialRequestByID(ctx, id)
	if err != nil {
		return err
	}

	// Normalisasi status ke Uppercase agar cocok dengan konstanta
	statusUpper := model.TrialRequestStatus(strings.ToUpper(string(status)))
	
	oldStatus := trial.Status
	trial.Status = statusUpper

	err = s.repo.Transaction(ctx, func(repo repository.SupportRepository) error {
		if err := repo.UpdateTrialRequest(ctx, trial); err != nil {
			return err
		}

		// If status changed to APPROVED, create a provisioning ticket
		if oldStatus != model.TrialRequestStatusApproved && statusUpper == model.TrialRequestStatusApproved {
			ticket := &model.ProvisioningTicket{
				TrialRequestID: trial.ID,
				Status:         model.ProvisioningTicketStatusWaiting,
			}
			if err := repo.CreateProvisioningTicket(ctx, ticket); err != nil {
				return err
			}
		}

		return nil
	})

	return err
}

func (s *supportService) GetAllSupportMessages(ctx context.Context) ([]modelDto.SupportMessageResponse, error) {
	messages, err := s.repo.FindAllSupportMessages(ctx, []string{"tenant", "user"})
	if err != nil {
		return nil, err
	}

	var responses []modelDto.SupportMessageResponse
	for _, m := range messages {
		responses = append(responses, mapToSupportMessageResponse(&m))
	}
	return responses, nil
}

func (s *supportService) UpdateSupportStatus(ctx context.Context, id uuid.UUID, status model.SupportStatus) error {
	msg, err := s.repo.FindSupportMessageByID(ctx, id, []string{})
	if err != nil {
		return err
	}

	msg.Status = status
	return s.repo.UpdateSupportMessage(ctx, msg)
}

func (s *supportService) GetAllProvisioningTickets(ctx context.Context) ([]modelDto.ProvisioningTicketResponse, error) {
	tickets, err := s.repo.FindAllProvisioningTickets(ctx, []string{"trial_request"})
	if err != nil {
		return nil, err
	}

	var responses []modelDto.ProvisioningTicketResponse
	for _, t := range tickets {
		responses = append(responses, mapToProvisioningTicketResponse(&t))
	}
	return responses, nil
}

func (s *supportService) ExecuteProvisioning(ctx context.Context, ticketID uuid.UUID, adminID uint) error {
	ticket, err := s.repo.FindProvisioningTicketByID(ctx, ticketID, []string{"trial_request"})
	if err != nil {
		return err
	}

	if ticket.Status == model.ProvisioningTicketStatusCompleted {
		return errors.New("ticket already completed")
	}

	if ticket.Status == model.ProvisioningTicketStatusExecuting {
		return errors.New("ticket currently executing")
	}

	ticket.Status = model.ProvisioningTicketStatusExecuting
	ticket.ExecutedBy = &adminID
	if err := s.repo.UpdateProvisioningTicket(ctx, ticket); err != nil {
		return err
	}

	// Logic Provisioning
	err = s.repo.Transaction(ctx, func(repo repository.SupportRepository) error {
		// 1. Create Tenant
		// Generate slug dari nama perusahaan + suffix random agar unik
		cleanName := strings.ToLower(strings.ReplaceAll(ticket.TrialRequest.CompanyName, " ", "-"))
		if len(cleanName) > 8 {
			cleanName = cleanName[:8]
		}
		tenantCode := fmt.Sprintf("%s-%s", cleanName, strings.ToLower(utils.GenerateRandomString(4)))

		tenant := &model.Tenant{
			Name: ticket.TrialRequest.CompanyName,
			Code: tenantCode,
		}
		if err := s.tenantRepo.Create(ctx, tenant); err != nil {
			return fmt.Errorf("failed to create tenant: %v", err)
		}

		// 2. Find ADMIN role (System Role)
		adminRole, err := s.roleRepo.FindByName(ctx, "admin")
		if err != nil {
			return fmt.Errorf("failed to find admin role: %v", err)
		}

		// 3. Setup Admin Account
		tempPassword := utils.GenerateRandomString(12)
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(tempPassword), bcrypt.DefaultCost)

		user := &model.User{
			Name:               ticket.TrialRequest.ContactName,
			Email:              ticket.TrialRequest.Email,
			Password:           string(hashedPassword),
			RoleID:             adminRole.ID,
			TenantID:           tenant.ID,
			EmployeeID:         fmt.Sprintf("OWNER-%03d-001", tenant.ID),
			IsSystemCreated:    true,
			MustChangePassword: true,
		}
		if err := s.userRepo.Create(ctx, user); err != nil {
			return fmt.Errorf("failed to create admin user: %v", err)
		}

		// 🆕 Automatic creation of Payroll Profile baseline
		profile := &model.UserPayrollProfile{
			UserID: user.ID,
		}
		if err := s.profileRepo.Upsert(ctx, profile); err != nil {
			return fmt.Errorf("failed to create user payroll profile: %v", err)
		}

		// 4. Create Default Subscription (Trial)
		subscription := &model.Subscription{
			TenantID:        tenant.ID,
			Plan:            "Basic", // Default plan
			BillingCycle:    model.BillingCycleMonthly,
			Amount:          0, // Trial is free
			Status:          model.SubscriptionStatusTrial,
			NextBillingDate: time.Now().AddDate(0, 1, 0), // 30 days trial
		}
		if err := s.subscriptionRepo.Create(ctx, subscription); err != nil {
			return fmt.Errorf("failed to create subscription: %v", err)
		}

		// 5. Create Default Tenant Settings
		tenantSetting := &model.TenantSetting{
			TenantID:           tenant.ID,
			MaxRadiusMeter:     100,
			AllowRemote:        true,
			RequireLocation:    true,
			ClockInStartTime:   "08:00",
			ClockInEndTime:     "09:00",
			ClockOutStartTime:  "17:00",
			ClockOutEndTime:    "18:00",
			LateAfterMinute:    15,
			RequireSelfie:      true,
			AllowMultipleCheck: false,
		}
		if err := s.tenantSettingRepo.Create(ctx, tenantSetting); err != nil {
			return fmt.Errorf("failed to create tenant settings: %v", err)
		}

		// 6. Dispatch Email
		emailHtml := utils.GetWelcomeEmailTemplate(user.Name, user.Email, tempPassword, tenant.Name, "")
		subject := fmt.Sprintf("Welcome to %s - Your Account Details", tenant.Name)
		go func() {
			emailCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			_ = utils.SendEmail(emailCtx, []string{user.Email}, subject, emailHtml)
		}()

		// 5. Mark Ticket COMPLETED
		now := time.Now()
		ticket.Status = model.ProvisioningTicketStatusCompleted
		ticket.CompletedAt = &now
		if err := repo.UpdateProvisioningTicket(ctx, ticket); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		ticket.Status = model.ProvisioningTicketStatusFailed
		ticket.ErrorLog = err.Error()
		_ = s.repo.UpdateProvisioningTicket(ctx, ticket)
		return err
	}

	return nil
}

func (s *supportService) CreateSupportMessage(ctx context.Context, tenantID uint, userID uint, req modelDto.CreateSupportMessageRequest) (modelDto.SupportMessageResponse, error) {
	message := &model.SupportMessage{
		TenantID: tenantID,
		UserID:   userID,
		Subject:  req.Subject,
		Message:  req.Message,
		Category: req.Category,
		Status:   model.SupportStatusPending,
	}

	if err := s.repo.CreateSupportMessage(ctx, message); err != nil {
		return modelDto.SupportMessageResponse{}, err
	}

	return mapToSupportMessageResponse(message), nil
}

// Helpers
func mapToTrialRequestResponse(t *model.TrialRequest) modelDto.TrialRequestResponse {
	return modelDto.TrialRequestResponse{
		ID:                 t.ID,
		CompanyName:        t.CompanyName,
		ContactName:        t.ContactName,
		Email:              t.Email,
		PhoneNumber:        t.PhoneNumber,
		EmployeeCountRange: t.EmployeeCountRange,
		Industry:           t.Industry,
		Status:             t.Status,
		CreatedAt:          t.CreatedAt,
	}
}

func mapToProvisioningTicketResponse(t *model.ProvisioningTicket) modelDto.ProvisioningTicketResponse {
	res := modelDto.ProvisioningTicketResponse{
		ID:             t.ID,
		TrialRequestID: t.TrialRequestID,
		Status:         t.Status,
		ErrorLog:       t.ErrorLog,
		ExecutedBy:     t.ExecutedBy,
		CompletedAt:    t.CompletedAt,
		CreatedAt:      t.CreatedAt,
	}
	if t.TrialRequest.ID != uuid.Nil {
		tr := mapToTrialRequestResponse(&t.TrialRequest)
		res.TrialRequest = &tr
	}
	return res
}

func mapToSupportMessageResponse(m *model.SupportMessage) modelDto.SupportMessageResponse {
	res := modelDto.SupportMessageResponse{
		ID:        m.ID,
		TenantID:  m.TenantID,
		UserID:    m.UserID,
		Subject:   m.Subject,
		Message:   m.Message,
		Category:  m.Category,
		Status:    m.Status,
		CreatedAt: m.CreatedAt,
	}
	if m.Tenant.ID != 0 {
		res.Tenant = &model.TenantResponse{
			ID:   m.Tenant.ID,
			Name: m.Tenant.Name,
		}
	}
	if m.User.ID != 0 {
		res.User = &model.UserResponse{
			ID:    m.User.ID,
			Name:  m.User.Name,
			Email: m.User.Email,
		}
	}
	return res
}
