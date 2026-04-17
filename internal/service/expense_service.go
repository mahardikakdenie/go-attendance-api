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

type ExpenseService interface {
	GetAllExpenses(ctx context.Context, filter model.ExpenseFilter) ([]dto.ExpenseResponse, int64, error)
	GetSummary(ctx context.Context, tenantID uint) (dto.ExpenseSummaryResponse, error)
	SubmitExpense(ctx context.Context, userID, tenantID uint, req dto.CreateExpenseRequest) (dto.ExpenseResponse, error)
	ApproveExpense(ctx context.Context, id uint, adminID uint) error
	RejectExpense(ctx context.Context, id uint, adminID uint, reason string) error
	UpdateQuota(ctx context.Context, userID uint, quota float64, adminID uint) error
}

type expenseService struct {
	repo         repository.ExpenseRepository
	userRepo     repository.UserRepository
	activityRepo repository.RecentActivityRepository
}

func NewExpenseService(repo repository.ExpenseRepository, userRepo repository.UserRepository, activityRepo repository.RecentActivityRepository) ExpenseService {
	return &expenseService{repo: repo, userRepo: userRepo, activityRepo: activityRepo}
}

func (s *expenseService) GetAllExpenses(ctx context.Context, filter model.ExpenseFilter) ([]dto.ExpenseResponse, int64, error) {
	expenses, total, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	var res []dto.ExpenseResponse
	for _, e := range expenses {
		res = append(res, s.mapToResponse(&e))
	}

	return res, total, nil
}

func (s *expenseService) GetSummary(ctx context.Context, tenantID uint) (dto.ExpenseSummaryResponse, error) {
	pending, approved, topCat, topPct, err := s.repo.GetSummary(ctx, tenantID)
	if err != nil {
		return dto.ExpenseSummaryResponse{}, err
	}

	return dto.ExpenseSummaryResponse{
		PendingAmount:           pending,
		ApprovedThisMonthAmount: approved,
		TopCategory: dto.ExpenseTopCategory{
			Name:       topCat,
			Percentage: topPct,
		},
	}, nil
}

func (s *expenseService) SubmitExpense(ctx context.Context, userID, tenantID uint, req dto.CreateExpenseRequest) (dto.ExpenseResponse, error) {
	// 1. Get User Quota
	user, err := s.userRepo.FindByID(ctx, userID, []string{})
	if err != nil {
		return dto.ExpenseResponse{}, err
	}

	// 2. Calculate current month usage (Approved + Pending)
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.Local)
	
	expenses, _, _ := s.repo.FindAll(ctx, model.ExpenseFilter{
		TenantID: tenantID,
		UserID:   userID,
	})

	var currentUsage float64
	for _, e := range expenses {
		if (e.Status == model.ExpenseStatusApproved || e.Status == model.ExpenseStatusPending) && e.Date.After(startOfMonth.AddDate(0, 0, -1)) {
			currentUsage += e.Amount
		}
	}

	if currentUsage+req.Amount > user.ExpenseQuota {
		return dto.ExpenseResponse{}, fmt.Errorf("quota tidak cukup. Sisa kuota bulan ini: %.2f", user.ExpenseQuota-currentUsage)
	}

	date, _ := time.Parse("2006-01-02", req.Date)

	expense := &model.Expense{
		TenantID:    tenantID,
		UserID:      userID,
		Category:    req.Category,
		Amount:      req.Amount,
		Date:        date,
		Description: req.Description,
		ReceiptUrl:  req.Receipt,
		Status:      model.ExpenseStatusPending,
	}

	if err := s.repo.Create(ctx, expense); err != nil {
		return dto.ExpenseResponse{}, err
	}

	// Log Activity
	_ = s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: userID,
		Title:  "Finance",
		Action: fmt.Sprintf("Submitted expense claim for %v", req.Amount),
		Status: "success",
	})

	// Re-load with user info
	e, _ := s.repo.FindByID(ctx, expense.ID)
	return s.mapToResponse(e), nil
}

func (s *expenseService) UpdateQuota(ctx context.Context, userID uint, quota float64, adminID uint) error {
	err := s.userRepo.UpdateQuota(ctx, userID, quota)
	if err != nil {
		return err
	}

	// Log Activity
	_ = s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: adminID,
		Title:  "Finance",
		Action: fmt.Sprintf("Updated expense quota for User ID %d to %v", userID, quota),
		Status: "success",
	})

	return nil
}

func (s *expenseService) ApproveExpense(ctx context.Context, id uint, adminID uint) error {
	expense, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if expense.Status != model.ExpenseStatusPending {
		return errors.New("expense is no longer pending")
	}

	expense.Status = model.ExpenseStatusApproved
	if err := s.repo.Update(ctx, expense); err != nil {
		return err
	}

	// 📉 Decrease user quota
	if err := s.userRepo.DecreaseQuota(ctx, expense.UserID, expense.Amount); err != nil {
		// Log error but maybe don't fail the whole request if status already updated?
		// Ideally use transaction, but let's at least try to keep it consistent.
		fmt.Printf("Warning: Failed to decrease quota for user %d: %v\n", expense.UserID, err)
	}

	// Log Activity for Admin
	_ = s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: adminID,
		Title:  "Finance",
		Action: fmt.Sprintf("Approved expense claim ID %d", id),
		Status: "success",
	})

	return nil
}

func (s *expenseService) RejectExpense(ctx context.Context, id uint, adminID uint, reason string) error {
	expense, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	if expense.Status != model.ExpenseStatusPending {
		return errors.New("expense is no longer pending")
	}

	expense.Status = model.ExpenseStatusRejected
	expense.AdminNotes = reason
	if err := s.repo.Update(ctx, expense); err != nil {
		return err
	}

	// Log Activity for Admin
	_ = s.activityRepo.Create(ctx, &model.RecentActivity{
		UserID: adminID,
		Title:  "Finance",
		Action: fmt.Sprintf("Rejected expense claim ID %d", id),
		Status: "success",
	})

	return nil
}

func (s *expenseService) mapToResponse(e *model.Expense) dto.ExpenseResponse {
	return dto.ExpenseResponse{
		ID:           e.ID,
		ClaimID:      fmt.Sprintf("EXP-%03d", e.ID),
		EmployeeName: e.User.Name,
		Avatar:       e.User.MediaUrl,
		Category:     e.Category,
		Amount:       e.Amount,
		Date:         e.Date.Format("2006-01-02"),
		Description:  e.Description,
		Status:       e.Status,
		ReceiptUrl:   e.ReceiptUrl,
		AdminNotes:   e.AdminNotes,
	}
}
