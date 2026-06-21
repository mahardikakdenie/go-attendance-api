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

	GetAllSupportMessages(ctx context.Context, filter model.SupportMessageFilter) ([]modelDto.SupportMessageResponse, int64, error)
	GetUserSupportHistory(ctx context.Context, userID uint, filter model.SupportMessageFilter) ([]modelDto.UserSupportHistoryResponse, int64, error)
	GetSupportAgents(ctx context.Context, search string, limit, offset int) ([]model.UserResponse, int64, error)
	UpdateSupportStatus(ctx context.Context, id uuid.UUID, status model.SupportStatus) error
	BulkUpdateSupportMessages(ctx context.Context, req modelDto.BulkSupportInboxRequest) error
	BulkAssignSupport(ctx context.Context, req modelDto.BulkAssignSupportRequest) (modelDto.BulkAssignResponse, error)
	UpdateSupportReadState(ctx context.Context, id uuid.UUID, isRead bool) error
	AssignSupportAgent(ctx context.Context, id uuid.UUID, agentID uint) (modelDto.SupportMessageResponse, error)
	ReplyToTicketByUser(ctx context.Context, userID uint, ticketID uuid.UUID, message string) (modelDto.SupportReplyResponse, error)

	// Superadmin Only
	GetAllProvisioningTickets(ctx context.Context) ([]modelDto.ProvisioningTicketResponse, error)
	ExecuteProvisioning(ctx context.Context, ticketID uuid.UUID, adminID uint) error

	// Tenant User
	CreateSupportMessage(ctx context.Context, tenantID uint, userID uint, req modelDto.CreateSupportMessageRequest) (modelDto.SupportMessageResponse, error)
	CreateReply(ctx context.Context, userID uint, messageID uuid.UUID, message string) (modelDto.SupportReplyResponse, error)
	GetReplies(ctx context.Context, messageID uuid.UUID) ([]modelDto.SupportReplyResponse, error)
	GetSupportCategories(ctx context.Context) []modelDto.SupportCategoryInfo
	GetSupportPriorities(ctx context.Context) []modelDto.SupportPriorityInfo
}

type supportService struct {
	repo              repository.SupportRepository
	superadminRepo    repository.SuperadminRepository
	tenantRepo        repository.TenantRepository
	userRepo          repository.UserRepository
	roleRepo          repository.RoleRepository
	subscriptionRepo  repository.SubscriptionRepository
	tenantSettingRepo repository.TenantSettingRepository
	profileRepo       repository.UserPayrollProfileRepository
	notifService      NotificationService
}

func NewSupportService(
	repo repository.SupportRepository,
	superadminRepo repository.SuperadminRepository,
	tenantRepo repository.TenantRepository,
	userRepo repository.UserRepository,
	roleRepo repository.RoleRepository,
	subscriptionRepo repository.SubscriptionRepository,
	tenantSettingRepo repository.TenantSettingRepository,
	profileRepo repository.UserPayrollProfileRepository,
	notifService NotificationService,
) SupportService {
	return &supportService{
		repo:              repo,
		superadminRepo:    superadminRepo,
		tenantRepo:        tenantRepo,
		userRepo:          userRepo,
		roleRepo:          roleRepo,
		subscriptionRepo:  subscriptionRepo,
		tenantSettingRepo: tenantSettingRepo,
		profileRepo:       profileRepo,
		notifService:      notifService,
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

	// 🆕 NOTIFICATION: Notify all Superadmins
	superadmins, _, _ := s.userRepo.FindAll(ctx, model.UserFilter{BaseRole: model.BaseRoleSuperAdmin}, nil)
	for _, sa := range superadmins {
		s.notifService.SendNotification(ctx, sa.TenantID, sa.ID, "New Trial Request", fmt.Sprintf("New trial request from %s (%s)", trial.CompanyName, trial.ContactName), model.NotificationTypeSupport)
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

func (s *supportService) GetAllSupportMessages(ctx context.Context, filter model.SupportMessageFilter) ([]modelDto.SupportMessageResponse, int64, error) {
	messages, total, err := s.repo.FindAllSupportMessages(ctx, filter, []string{"tenant", "user", "assigned_to"})
	if err != nil {
		return nil, 0, err
	}

	var responses []modelDto.SupportMessageResponse
	for _, m := range messages {
		responses = append(responses, mapToSupportMessageResponse(&m))
	}
	return responses, total, nil
}

func (s *supportService) toUserResponse(user model.User) model.UserResponse {
	resp := model.UserResponse{
		ID:        user.ID,
		Name:      user.Name,
		Email:     user.Email,
		TenantID:  user.TenantID,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		BaseRole:  user.Role.BaseRole,
	}
	if user.Role.ID != 0 {
		resp.Role = &model.RoleResponse{
			ID:       user.Role.ID,
			Name:     user.Role.Name,
			BaseRole: user.Role.BaseRole,
		}
	}
	return resp
}

func (s *supportService) GetSupportAgents(ctx context.Context, search string, limit, offset int) ([]model.UserResponse, int64, error) {
	if limit <= 0 {
		limit = 10
	}

	users, total, err := s.superadminRepo.GetPlatformAccounts(ctx, search, "", limit, offset)
	if err != nil {
		return nil, 0, err
	}

	// ISSUE-002: Pre-allocate with capacity and ensure it's not nil
	responses := make([]model.UserResponse, 0, len(users))
	for _, user := range users {
		responses = append(responses, s.toUserResponse(user))
	}

	return responses, total, nil
}

func (s *supportService) BulkAssignSupport(ctx context.Context, req modelDto.BulkAssignSupportRequest) (modelDto.BulkAssignResponse, error) {
	if len(req.IDs) == 0 {
		return modelDto.BulkAssignResponse{}, errors.New("no ticket ids provided")
	}
	if req.AgentID == 0 {
		return modelDto.BulkAssignResponse{}, errors.New("invalid agent_id")
	}

	// Verify agent exists and belongs to HQ
	agent, err := s.userRepo.FindByID(ctx, req.AgentID, nil)
	if err != nil {
		return modelDto.BulkAssignResponse{}, fmt.Errorf("agent user not found: %w", err)
	}
	if agent.TenantID != 1 {
		return modelDto.BulkAssignResponse{}, errors.New("can only assign tickets to HQ members (Tenant 1)")
	}

	updates := map[string]interface{}{
		"assigned_to_id": req.AgentID,
	}

	if err := s.repo.BulkUpdateSupportMessages(ctx, req.IDs, updates); err != nil {
		return modelDto.BulkAssignResponse{}, err
	}

	// Notify assigned agent once for all tickets
	title := "Support Tickets Assigned"
	notifMessage := fmt.Sprintf("You have been assigned %d support ticket(s).", len(req.IDs))
	s.notifService.SendNotification(ctx, agent.TenantID, agent.ID, title, notifMessage, model.NotificationTypeSupport)

	// Email agent (async)
	go func(agentName, agentEmail string, count int) {
		emailCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		subject := fmt.Sprintf("[Support Desk] %d Ticket(s) Assigned to You", count)
		html := utils.GetSupportTicketAssignedEmailTemplate(agentName, fmt.Sprintf("%d tickets assigned to you", count), "Multiple Tenants")
		_ = utils.SendEmail(emailCtx, []string{agentEmail}, subject, html)
	}(agent.Name, agent.Email, len(req.IDs))

	return modelDto.BulkAssignResponse{
		Updated: len(req.IDs),
		Failed:  0,
	}, nil
}

func (s *supportService) BulkUpdateSupportMessages(ctx context.Context, req modelDto.BulkSupportInboxRequest) error {
	if len(req.IDs) == 0 {
		return errors.New("no ticket ids provided")
	}

	updates := make(map[string]interface{})
	switch req.Action {
	case "MARK_READ":
		updates["is_read"] = true
	case "MARK_UNREAD":
		updates["is_read"] = false
	case "RESOLVE":
		updates["status"] = model.SupportStatusResolved
	case "ASSIGN":
		if req.AssignToID == nil {
			return errors.New("assign_to_id is required for ASSIGN action")
		}
		// Verify agent exists and belongs to HQ
		agent, err := s.userRepo.FindByID(ctx, *req.AssignToID, nil)
		if err != nil {
			return fmt.Errorf("agent user not found: %w", err)
		}
		if agent.TenantID != 1 {
			return errors.New("can only assign tickets to HQ members (Tenant 1)")
		}
		updates["assigned_to_id"] = *req.AssignToID
	default:
		return errors.New("invalid action")
	}

	return s.repo.BulkUpdateSupportMessages(ctx, req.IDs, updates)
}

func (s *supportService) UpdateSupportReadState(ctx context.Context, id uuid.UUID, isRead bool) error {
	msg, err := s.repo.FindSupportMessageByID(ctx, id, []string{})
	if err != nil {
		return err
	}
	msg.IsRead = isRead
	return s.repo.UpdateSupportMessage(ctx, msg)
}

func (s *supportService) AssignSupportAgent(ctx context.Context, id uuid.UUID, agentID uint) (modelDto.SupportMessageResponse, error) {
	msg, err := s.repo.FindSupportMessageByID(ctx, id, []string{"tenant", "user", "assigned_to"})
	if err != nil {
		return modelDto.SupportMessageResponse{}, err
	}

	// Verify agent exists and belongs to HQ
	agent, err := s.userRepo.FindByID(ctx, agentID, nil)
	if err != nil {
		return modelDto.SupportMessageResponse{}, fmt.Errorf("agent user not found: %w", err)
	}
	if agent.TenantID != 1 {
		return modelDto.SupportMessageResponse{}, errors.New("can only assign tickets to HQ members (Tenant 1)")
	}

	msg.AssignedToID = &agentID

	// Temporarily attach agent object so mapToSupportMessageResponse includes it immediately
	msg.AssignedTo = agent

	if err := s.repo.UpdateSupportMessage(ctx, msg); err != nil {
		return modelDto.SupportMessageResponse{}, err
	}

	// Notify assigned agent in-app
	title := "New Support Assignment"
	notifMessage := fmt.Sprintf("You are assigned ticket: %s", msg.Subject)
	s.notifService.SendNotification(ctx, agent.TenantID, agent.ID, title, notifMessage, model.NotificationTypeSupport)

	// Send email to assigned agent (async, non-blocking)
	go func(agentName, agentEmail, subject, tenantName string) {
		emailCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		html := utils.GetSupportTicketAssignedEmailTemplate(agentName, subject, tenantName)
		_ = utils.SendEmail(emailCtx, []string{agentEmail}, "[Support Desk] New Ticket Assigned", html)
	}(agent.Name, agent.Email, msg.Subject, msg.Tenant.Name)

	return mapToSupportMessageResponse(msg), nil
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
		trialPlan, err := s.subscriptionRepo.FindPlanByName(ctx, "Trial")
		if err != nil {
			return fmt.Errorf("failed to find trial plan: %v", err)
		}

		subscription := &model.Subscription{
			TenantID:        tenant.ID,
			PlanID:          trialPlan.ID,
			BillingCycle:    model.BillingCycleMonthly,
			Amount:          0, // Trial is free
			Status:          model.SubscriptionStatusTrial,
			NextBillingDate: utils.Now().AddDate(0, 0, 14), // 14 days trial
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
		now := utils.Now()
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
	priority := req.Priority
	if priority == "" {
		priority = model.SupportPriorityMedium
	}

	message := &model.SupportMessage{
		TenantID:      tenantID,
		UserID:        userID,
		Subject:       req.Subject,
		Message:       req.Message,
		Category:      req.Category,
		Priority:      priority,
		AttachmentURL: req.AttachmentURL,
		Status:        model.SupportStatusPending,
	}

	if err := s.repo.CreateSupportMessage(ctx, message); err != nil {
		return modelDto.SupportMessageResponse{}, err
	}

	// 🆕 NOTIFICATION: Notify HQ Admins (Tenant 1)
	hqAdmins, _, _ := s.userRepo.FindAll(ctx, model.UserFilter{TenantID: 1}, nil)
	for _, admin := range hqAdmins {
		if admin.Role != nil && admin.Role.Name == "admin" {
			s.notifService.SendNotification(ctx, 1, admin.ID, "New Support Ticket", fmt.Sprintf("New ticket from Tenant %d: %s", tenantID, message.Subject), model.NotificationTypeSupport)
		}
	}

	return mapToSupportMessageResponse(message), nil
}

func (s *supportService) CreateReply(ctx context.Context, userID uint, messageID uuid.UUID, message string) (modelDto.SupportReplyResponse, error) {
	ticket, err := s.repo.FindSupportMessageByID(ctx, messageID, []string{"user"})
	if err != nil {
		return modelDto.SupportReplyResponse{}, err
	}

	reply := &model.SupportReply{
		MessageID: messageID,
		UserID:    userID,
		Message:   message,
	}

	if err := s.repo.CreateReply(ctx, reply); err != nil {
		return modelDto.SupportReplyResponse{}, err
	}

	// 🆕 NOTIFICATION: Two-way
	if userID == ticket.UserID {
		// User replied, notify HQ Admins
		hqAdmins, _, _ := s.userRepo.FindAll(ctx, model.UserFilter{TenantID: 1}, nil)
		for _, admin := range hqAdmins {
			if admin.Role != nil && admin.Role.Name == "admin" {
				s.notifService.SendNotification(ctx, 1, admin.ID, "Support Ticket Reply", fmt.Sprintf("User replied to ticket: %s", ticket.Subject), model.NotificationTypeSupport)
			}
		}
	} else {
		// Admin replied, notify the user who opened the ticket
		s.notifService.SendNotification(ctx, ticket.TenantID, ticket.UserID, "Support Response", fmt.Sprintf("Support has replied to your ticket: %s", ticket.Subject), model.NotificationTypeSupport)

		// Auto-update status to IN_PROGRESS if it was PENDING
		if ticket.Status == model.SupportStatusPending {
			ticket.Status = model.SupportStatusInProgress
			_ = s.repo.UpdateSupportMessage(ctx, ticket)
		}
	}

	return mapToSupportReplyResponse(reply), nil
}

func (s *supportService) GetReplies(ctx context.Context, messageID uuid.UUID) ([]modelDto.SupportReplyResponse, error) {
	replies, err := s.repo.FindRepliesByMessageID(ctx, messageID)
	if err != nil {
		return nil, err
	}

	var responses []modelDto.SupportReplyResponse
	for _, r := range replies {
		responses = append(responses, mapToSupportReplyResponse(&r))
	}
	return responses, nil
}

func (s *supportService) GetUserSupportHistory(ctx context.Context, userID uint, filter model.SupportMessageFilter) ([]modelDto.UserSupportHistoryResponse, int64, error) {
	messages, total, err := s.repo.FindAllUserSupportMessages(ctx, userID, filter, []string{})
	if err != nil {
		return nil, 0, err
	}

	var responses []modelDto.UserSupportHistoryResponse
	for _, m := range messages {
		history := modelDto.UserSupportHistoryResponse{
			ID:        m.ID,
			Subject:   m.Subject,
			Category:  m.Category,
			Priority:  m.Priority,
			Status:    m.Status,
			Message:   m.Message,
			CreatedAt: m.CreatedAt,
			Replies:   make([]modelDto.UserSupportReplyItemResponse, 0, len(m.Replies)),
		}

		for _, r := range m.Replies {
			senderType := "USER"
			if r.User.TenantID == 1 {
				senderType = "ADMIN"
			}
			history.Replies = append(history.Replies, modelDto.UserSupportReplyItemResponse{
				ID:         r.ID,
				SenderType: senderType,
				SenderName: r.User.Name,
				Message:    r.Message,
				CreatedAt:  r.CreatedAt,
			})
		}

		responses = append(responses, history)
	}
	return responses, total, nil
}

func (s *supportService) ReplyToTicketByUser(ctx context.Context, userID uint, ticketID uuid.UUID, message string) (modelDto.SupportReplyResponse, error) {
	ticket, err := s.repo.FindSupportMessageByID(ctx, ticketID, []string{"user"})
	if err != nil {
		return modelDto.SupportReplyResponse{}, err
	}

	// Validate ownership
	if ticket.UserID != userID {
		return modelDto.SupportReplyResponse{}, errors.New("you do not have permission to reply to this ticket")
	}

	// Validate status
	if ticket.Status == model.SupportStatusClosed {
		return modelDto.SupportReplyResponse{}, errors.New("this ticket has been closed and cannot be replied to")
	}

	reply := &model.SupportReply{
		MessageID: ticketID,
		UserID:    userID,
		Message:   message,
	}

	if err := s.repo.CreateReply(ctx, reply); err != nil {
		return modelDto.SupportReplyResponse{}, err
	}

	// Update status if RESOLVED -> PENDING
	if ticket.Status == model.SupportStatusResolved {
		ticket.Status = model.SupportStatusPending
		_ = s.repo.UpdateSupportMessage(ctx, ticket)
	}

	// Notify HQ Admins (Tenant 1) that user replied
	hqAdmins, _, _ := s.userRepo.FindAll(ctx, model.UserFilter{TenantID: 1}, nil)
	for _, admin := range hqAdmins {
		if admin.Role != nil && (admin.Role.BaseRole == model.BaseRoleAdmin || admin.Role.BaseRole == model.BaseRoleSuperAdmin) {
			s.notifService.SendNotification(ctx, 1, admin.ID, "Support Ticket Reply", fmt.Sprintf("User replied to ticket: %s", ticket.Subject), model.NotificationTypeSupport)
		}
	}

	return mapToSupportReplyResponse(reply), nil
}

func (s *supportService) GetSupportCategories(ctx context.Context) []modelDto.SupportCategoryInfo {
	return []modelDto.SupportCategoryInfo{
		{ID: string(model.SupportCategoryTechnical), Label: "Technical/Bug"},
		{ID: string(model.SupportCategoryBilling), Label: "Billing & Subscription"},
		{ID: string(model.SupportCategoryFeature), Label: "Feature Request"},
		{ID: string(model.SupportCategoryAccount), Label: "Account & Login"},
		{ID: string(model.SupportCategoryIntegration), Label: "Integration/API"},
		{ID: string(model.SupportCategoryOther), Label: "Other"},
	}
}

func (s *supportService) GetSupportPriorities(ctx context.Context) []modelDto.SupportPriorityInfo {
	return []modelDto.SupportPriorityInfo{
		{ID: string(model.SupportPriorityLow), Label: "Low", Color: "gray"},
		{ID: string(model.SupportPriorityMedium), Label: "Medium", Color: "blue"},
		{ID: string(model.SupportPriorityHigh), Label: "High", Color: "orange"},
		{ID: string(model.SupportPriorityUrgent), Label: "Urgent", Color: "red"},
	}
}

// Helpers
func mapToSupportReplyResponse(r *model.SupportReply) modelDto.SupportReplyResponse {
	res := modelDto.SupportReplyResponse{
		ID:        r.ID,
		MessageID: r.MessageID,
		UserID:    r.UserID,
		Message:   r.Message,
		CreatedAt: r.CreatedAt,
	}
	if r.User.ID != 0 {
		res.User = &model.UserResponse{
			ID:    r.User.ID,
			Name:  r.User.Name,
			Email: r.User.Email,
		}
	}
	return res
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
		ID:            m.ID,
		TenantID:      m.TenantID,
		UserID:        m.UserID,
		Subject:       m.Subject,
		Message:       m.Message,
		Category:      m.Category,
		Priority:      m.Priority,
		Status:        m.Status,
		IsRead:        m.IsRead,
		AttachmentURL: m.AttachmentURL,
		CreatedAt:     m.CreatedAt,
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
	if m.AssignedToID != nil && m.AssignedTo != nil && m.AssignedTo.ID != 0 {
		res.AssignedTo = &modelDto.SupportAssigneeResponse{
			ID:   m.AssignedTo.ID,
			Name: m.AssignedTo.Name,
		}
	}
	return res
}
