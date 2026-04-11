package service

import (
	"context"
	"errors"
	"fmt"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/utils"
	"time"
)

type LeaveService interface {
	RequestLeave(ctx context.Context, userID uint, tenantID uint, req model.LeaveRequest) (model.LeaveResponse, error)
	GetLeaveBalances(ctx context.Context, userID uint) ([]model.LeaveBalance, error)
	GetLeaveHistory(ctx context.Context, userID uint, limit, offset int) ([]model.LeaveResponse, int64, error)
	GetPendingCount(ctx context.Context, userID uint) (int, error)
}

type leaveService struct {
	repo         repository.LeaveRepository
	activityRepo repository.RecentActivityRepository
	userRepo     repository.UserRepository
	orgService   OrganizationService
}

func NewLeaveService(
	repo repository.LeaveRepository,
	activityRepo repository.RecentActivityRepository,
	userRepo repository.UserRepository,
	orgService OrganizationService,
) LeaveService {
	return &leaveService{
		repo:         repo,
		activityRepo: activityRepo,
		userRepo:     userRepo,
		orgService:   orgService,
	}
}

func (s *leaveService) GetPendingCount(ctx context.Context, userID uint) (int, error) {
	count, err := s.repo.GetPendingCount(ctx, userID)
	return int(count), err
}

func (s *leaveService) RequestLeave(ctx context.Context, userID uint, tenantID uint, req model.LeaveRequest) (model.LeaveResponse, error) {
	startDate, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		return model.LeaveResponse{}, errors.New("invalid start date format")
	}
	endDate, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		return model.LeaveResponse{}, errors.New("invalid end date format")
	}

	if endDate.Before(startDate) {
		return model.LeaveResponse{}, errors.New("end date cannot be before start date")
	}

	// Calculate total days
	totalDays := int(endDate.Sub(startDate).Hours()/24) + 1

	// Check balance
	year := startDate.Year()
	balance, err := s.repo.GetBalance(ctx, userID, req.LeaveTypeID, year)
	if err != nil {
		return model.LeaveResponse{}, err
	}

	fmt.Println("Balance ", balance)

	if balance == nil || balance.Balance < totalDays {
		return model.LeaveResponse{}, errors.New("insufficient leave balance")
	}

	// Get User info
	user, _ := s.userRepo.FindByID(ctx, userID, nil)

	// Get the correct manager for approval (with escalation if on leave)
	approvalManager, _ := s.orgService.GetApprovalManager(ctx, userID, startDate)

	leave := &model.Leave{
		TenantID:    tenantID,
		UserID:      userID,
		LeaveTypeID: req.LeaveTypeID,
		StartDate:   startDate,
		EndDate:     endDate,
		Reason:      req.Reason,
		Status:      model.LeaveStatusPending,
		DelegateID:  req.DelegateID,
	}

	if err := s.repo.CreateLeave(ctx, leave); err != nil {
		return model.LeaveResponse{}, err
	}

	// Update User's Delegate for this period (Simple implementation)
	if req.DelegateID != nil && user != nil {
		user.DelegateID = req.DelegateID
		s.userRepo.Update(ctx, user)
	}

	// Update balance (deduct)
	balance.Balance -= totalDays
	if err := s.repo.UpdateBalance(ctx, balance); err != nil {
		return model.LeaveResponse{}, err
	}

	lt, _ := s.repo.GetLeaveTypeByID(ctx, req.LeaveTypeID)

	// Record activity
	s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: userID,
		Title:  fmt.Sprintf("Requested %s for %d days", lt.Name, totalDays),
		Action: "Leave Request",
		Status: "Pending",
	})

	// NOTIFICATION LOGIC
	// 1. Notify Approval Manager (Could be direct manager or escalated)
	if approvalManager != nil && user != nil {
		subject := fmt.Sprintf("Leave Approval Needed: %s", user.Name)
		html := utils.GetLeaveApprovalRequestTemplate(
			approvalManager.Name,
			user.Name,
			lt.Name,
			req.StartDate,
			req.EndDate,
			totalDays,
			req.Reason,
		)
		
		utils.SendEmail([]string{approvalManager.Email}, subject, html)
	}

	// 2. Notify Delegate
	if req.DelegateID != nil {
		delegate, _ := s.userRepo.FindByID(ctx, *req.DelegateID, nil)
		if delegate != nil {
			subject := fmt.Sprintf("Leave Delegation: %s", user.Name)
			html := utils.GetLeaveDelegationTemplate(
				delegate.Name,
				user.Name,
				req.StartDate,
				req.EndDate,
			)
			
			utils.SendEmail([]string{delegate.Email}, subject, html)
		}
	}

	return model.LeaveResponse{
		ID:          leave.ID,
		UserID:      leave.UserID,
		LeaveTypeID: leave.LeaveTypeID,
		LeaveType:   lt.Name,
		StartDate:   leave.StartDate,
		EndDate:     leave.EndDate,
		TotalDays:   totalDays,
		Reason:      leave.Reason,
		Status:      leave.Status,
		CreatedAt:   leave.CreatedAt,
	}, nil
}

func (s *leaveService) GetLeaveBalances(ctx context.Context, userID uint) ([]model.LeaveBalance, error) {
	year := time.Now().Year()
	balances, err := s.repo.GetBalancesByUser(ctx, userID, year)
	if err != nil {
		return nil, err
	}
	return balances, nil
}

func (s *leaveService) GetLeaveHistory(ctx context.Context, userID uint, limit, offset int) ([]model.LeaveResponse, int64, error) {
	leaves, total, err := s.repo.GetLeavesByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var responses []model.LeaveResponse
	for _, l := range leaves {
		totalDays := int(l.EndDate.Sub(l.StartDate).Hours()/24) + 1
		responses = append(responses, model.LeaveResponse{
			ID:          l.ID,
			UserID:      l.UserID,
			LeaveTypeID: l.LeaveTypeID,
			LeaveType:   l.LeaveType.Name,
			StartDate:   l.StartDate,
			EndDate:     l.EndDate,
			TotalDays:   totalDays,
			Reason:      l.Reason,
			Status:      l.Status,
			CreatedAt:   l.CreatedAt,
		})
	}

	return responses, total, nil
}
