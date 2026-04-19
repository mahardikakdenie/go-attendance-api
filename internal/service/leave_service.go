package service

import (
	"context"
	"errors"
	"fmt"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/utils"
	"time"

	"github.com/redis/go-redis/v9"
)

type LeaveService interface {
	RequestLeave(ctx context.Context, userID uint, tenantID uint, req model.LeaveRequest) (model.LeaveResponse, error)
	GetLeaveBalances(ctx context.Context, userID uint) ([]model.LeaveBalance, error)
	GetLeaveHistory(ctx context.Context, requesterID uint, filter model.LeaveFilter, limit, offset int) ([]model.LeaveResponse, int64, error)
	GetPendingCount(ctx context.Context, userID uint) (int, error)
	ApproveLeave(ctx context.Context, approverID uint, leaveID uint, notes string) error
	RejectLeave(ctx context.Context, approverID uint, leaveID uint, notes string) error
}

type leaveService struct {
	repo         repository.LeaveRepository
	activityRepo repository.RecentActivityRepository
	userRepo     repository.UserRepository
	orgService   OrganizationService
	userService  UserService
	redis        *redis.Client
}

func NewLeaveService(
	repo repository.LeaveRepository,
	activityRepo repository.RecentActivityRepository,
	userRepo repository.UserRepository,
	orgService OrganizationService,
	userService UserService,
	redis *redis.Client,
) LeaveService {
	return &leaveService{
		repo:         repo,
		activityRepo: activityRepo,
		userRepo:     userRepo,
		orgService:   orgService,
		userService:  userService,
		redis:        redis,
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

			utils.SendEmail(ctx, []string{approvalManager.Email}, subject, html)
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

			utils.SendEmail(ctx, []string{delegate.Email}, subject, html)
			}
			}
	return model.LeaveResponse{
		ID:          leave.ID,
		UserID:      leave.UserID,
		UserName:    user.Name,
		UserAvatar:  user.MediaUrl,
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

func (s *leaveService) GetLeaveHistory(ctx context.Context, requesterID uint, filter model.LeaveFilter, limit, offset int) ([]model.LeaveResponse, int64, error) {
	// Apply Hierarchical Scoping
	if requesterID != 0 {
		allowedRoleIDs, _ := s.userService.GetAllowedRoleIDs(ctx, requesterID)
		filter.AllowedRoleIDs = allowedRoleIDs
	}

	leaves, total, err := s.repo.FindAll(ctx, filter, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	var responses []model.LeaveResponse
	for _, l := range leaves {
		totalDays := int(l.EndDate.Sub(l.StartDate).Hours()/24) + 1
		userName := ""
		userAvatar := ""
		if l.User != nil {
			userName = l.User.Name
			userAvatar = l.User.MediaUrl
		}

		responses = append(responses, model.LeaveResponse{
			ID:          l.ID,
			UserID:      l.UserID,
			UserName:    userName,
			UserAvatar:  userAvatar,
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

func (s *leaveService) ApproveLeave(ctx context.Context, approverID uint, leaveID uint, notes string) error {
	leave, err := s.repo.FindByID(ctx, leaveID)
	if err != nil {
		return err
	}

	if leave.Status != model.LeaveStatusPending {
		return errors.New("leave is not in pending status")
	}

	// Hierarchical Check: Is requester a subordinate of approver?
	approver, _ := s.userRepo.FindByID(ctx, approverID, []string{"role"})
	if approver == nil {
		return errors.New("approver not found")
	}

	// Dual-nature Superadmin bypass
	if approver.Role == nil || approver.Role.BaseRole != model.BaseRoleAdmin {
		allowedRoleIDs, _ := s.userService.GetAllowedRoleIDs(ctx, approverID)
		isSubordinate := false
		for _, rID := range allowedRoleIDs {
			if rID == leave.User.RoleID {
				isSubordinate = true
				break
			}
		}
		if !isSubordinate {
			return errors.New("forbidden: you do not have authority over this employee")
		}
	}

	leave.Status = model.LeaveStatusApproved
	leave.AdminNotes = notes
	if err := s.repo.Update(ctx, leave); err != nil {
		return err
	}

	// Invalidate HR Dashboard Cache
	s.invalidateHrDashboardCache(ctx, leave.TenantID)

	// Log activity for both
	s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: approverID,
		Title:  "Leave Approval",
		Action: fmt.Sprintf("Approved leave for %s", leave.User.Name),
		Status: "success",
	})
	s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: leave.UserID,
		Title:  "Leave Approved",
		Action: fmt.Sprintf("Your %s request was approved", leave.LeaveType.Name),
		Status: "success",
	})

	return nil
}

func (s *leaveService) RejectLeave(ctx context.Context, approverID uint, leaveID uint, notes string) error {
	leave, err := s.repo.FindByID(ctx, leaveID)
	if err != nil {
		return err
	}

	if leave.Status != model.LeaveStatusPending {
		return errors.New("leave is not in pending status")
	}

	// Hierarchical Check
	approver, _ := s.userRepo.FindByID(ctx, approverID, []string{"role"})
	if approver == nil {
		return errors.New("approver not found")
	}

	if approver.Role == nil || approver.Role.BaseRole != model.BaseRoleAdmin {
		allowedRoleIDs, _ := s.userService.GetAllowedRoleIDs(ctx, approverID)
		isSubordinate := false
		for _, rID := range allowedRoleIDs {
			if rID == leave.User.RoleID {
				isSubordinate = true
				break
			}
		}
		if !isSubordinate {
			return errors.New("forbidden: you do not have authority over this employee")
		}
	}

	// Refund balance
	year := leave.StartDate.Year()
	balance, _ := s.repo.GetBalance(ctx, leave.UserID, leave.LeaveTypeID, year)
	if balance != nil {
		totalDays := int(leave.EndDate.Sub(leave.StartDate).Hours()/24) + 1
		balance.Balance += totalDays
		s.repo.UpdateBalance(ctx, balance)
	}

	leave.Status = model.LeaveStatusRejected
	leave.AdminNotes = notes
	if err := s.repo.Update(ctx, leave); err != nil {
		return err
	}

	// Invalidate HR Dashboard Cache
	s.invalidateHrDashboardCache(ctx, leave.TenantID)

	// Log activity
	s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: approverID,
		Title:  "Leave Rejection",
		Action: fmt.Sprintf("Rejected leave for %s", leave.User.Name),
		Status: "success",
	})
	s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: leave.UserID,
		Title:  "Leave Rejected",
		Action: fmt.Sprintf("Your %s request was rejected", leave.LeaveType.Name),
		Status: "rejected",
	})

	return nil
}

func (s *leaveService) invalidateHrDashboardCache(ctx context.Context, tenantID uint) {
	cacheKey := fmt.Sprintf("cache:dashboard:hr:%d", tenantID)
	s.redis.Del(ctx, cacheKey)
}
