package service

import (
	"context"
	"errors"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"time"
)

type LeaveService interface {
	RequestLeave(ctx context.Context, userID uint, tenantID uint, req model.LeaveRequest) (model.LeaveResponse, error)
	GetLeaveBalances(ctx context.Context, userID uint) ([]model.LeaveBalance, error)
	GetLeaveHistory(ctx context.Context, userID uint, limit, offset int) ([]model.LeaveResponse, int64, error)
	GetPendingCount(ctx context.Context, userID uint) (int, error)
}

type leaveService struct {
	repo repository.LeaveRepository
}

func NewLeaveService(repo repository.LeaveRepository) LeaveService {
	return &leaveService{repo: repo}
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

	if balance == nil || balance.Balance < totalDays {
		return model.LeaveResponse{}, errors.New("insufficient leave balance")
	}

	leave := &model.Leave{
		TenantID:    tenantID,
		UserID:      userID,
		LeaveTypeID: req.LeaveTypeID,
		StartDate:   startDate,
		EndDate:     endDate,
		Reason:      req.Reason,
		Status:      model.LeaveStatusPending,
	}

	if err := s.repo.CreateLeave(ctx, leave); err != nil {
		return model.LeaveResponse{}, err
	}

	// Update balance (deduct)
	balance.Balance -= totalDays
	if err := s.repo.UpdateBalance(ctx, balance); err != nil {
		return model.LeaveResponse{}, err
	}

	lt, _ := s.repo.GetLeaveTypeByID(ctx, req.LeaveTypeID)

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
	// Not implemented in repo yet, but usually we'd need a GetBalancesByUser
	// For now, let's assume we return empty or implement GetBalancesByUser in repo
	return nil, nil
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
