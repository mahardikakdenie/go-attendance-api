package service

import (
	"context"
	"errors"
	dto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"time"
)

type PerformanceService interface {
	// Goals
	GetMyGoals(ctx context.Context, userID uint) ([]dto.GoalResponse, error)
	GetUserGoals(ctx context.Context, requesterID, userID uint) ([]dto.GoalResponse, error)
	CreateGoal(ctx context.Context, tenantID uint, requesterID uint, req dto.CreateGoalRequest) (dto.GoalResponse, error)
	UpdateGoalProgress(ctx context.Context, requesterID uint, goalID uint, progress float64) error

	// Cycles & Appraisals
	GetAllCycles(ctx context.Context) ([]dto.CycleResponse, error)
	GetAppraisalsByCycle(ctx context.Context, cycleID uint) ([]dto.AppraisalResponse, error)
	SubmitSelfReview(ctx context.Context, userID uint, appraisalID uint, req dto.SubmitSelfReviewRequest) error

	// Hooks
	SyncAttendanceKPI(ctx context.Context, userID uint, lateCount int) error
}

type performanceService struct {
	repo     repository.PerformanceRepository
	userRepo repository.UserRepository
}

func NewPerformanceService(repo repository.PerformanceRepository, userRepo repository.UserRepository) PerformanceService {
	return &performanceService{repo: repo, userRepo: userRepo}
}

func (s *performanceService) GetMyGoals(ctx context.Context, userID uint) ([]dto.GoalResponse, error) {
	goals, err := s.repo.FindGoalsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	res := make([]dto.GoalResponse, 0)
	for _, g := range goals {
		res = append(res, s.mapGoalToResponse(g))
	}
	return res, nil
}

func (s *performanceService) GetUserGoals(ctx context.Context, requesterID, userID uint) ([]dto.GoalResponse, error) {
	// Check if requester is manager of the user or is admin/hr
	if err := s.checkManagerOrAdminAccess(ctx, requesterID, userID); err != nil {
		return nil, err
	}

	return s.GetMyGoals(ctx, userID)
}

func (s *performanceService) CreateGoal(ctx context.Context, tenantID uint, requesterID uint, req dto.CreateGoalRequest) (dto.GoalResponse, error) {
	// Only managers or admin/hr can create goals for users
	if err := s.checkManagerOrAdminAccess(ctx, requesterID, req.UserID); err != nil {
		return dto.GoalResponse{}, err
	}

	start, _ := time.Parse("2006-01-02", req.StartDate)
	end, _ := time.Parse("2006-01-02", req.EndDate)

	goal := &model.PerformanceGoal{
		TenantID:    tenantID,
		UserID:      req.UserID,
		Title:       req.Title,
		Description: req.Description,
		Type:        req.Type,
		TargetValue: req.TargetValue,
		Unit:        req.Unit,
		StartDate:   start,
		EndDate:     end,
		Status:      model.GoalStatusInProgress,
	}

	if err := s.repo.CreateGoal(ctx, goal); err != nil {
		return dto.GoalResponse{}, err
	}

	return s.mapGoalToResponse(*goal), nil
}

func (s *performanceService) UpdateGoalProgress(ctx context.Context, requesterID uint, goalID uint, progress float64) error {
	goal, err := s.repo.FindGoalByID(ctx, goalID)
	if err != nil {
		return err
	}

	// Owner can update progress, or manager/admin
	if goal.UserID != requesterID {
		if err := s.checkManagerOrAdminAccess(ctx, requesterID, goal.UserID); err != nil {
			return err
		}
	}

	goal.CurrentProgress = progress
	if goal.CurrentProgress >= goal.TargetValue {
		goal.Status = model.GoalStatusCompleted
	}

	return s.repo.UpdateGoal(ctx, goal)
}

func (s *performanceService) GetAllCycles(ctx context.Context) ([]dto.CycleResponse, error) {
	cycles, err := s.repo.FindAllCycles(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]dto.CycleResponse, 0)
	for _, c := range cycles {
		res = append(res, dto.CycleResponse{
			ID:        c.ID,
			Name:      c.Name,
			StartDate: c.StartDate.Format("2006-01-02"),
			EndDate:   c.EndDate.Format("2006-01-02"),
			Status:    c.Status,
		})
	}
	return res, nil
}

func (s *performanceService) GetAppraisalsByCycle(ctx context.Context, cycleID uint) ([]dto.AppraisalResponse, error) {
	appraisals, err := s.repo.FindAppraisalsByCycleID(ctx, cycleID)
	if err != nil {
		return nil, err
	}

	res := make([]dto.AppraisalResponse, 0)
	for _, a := range appraisals {
		res = append(res, dto.AppraisalResponse{
			ID:           a.ID,
			CycleID:      a.CycleID,
			UserID:       a.UserID,
			UserName:     a.User.Name,
			SelfScore:    a.SelfScore,
			ManagerScore: a.ManagerScore,
			FinalScore:   a.FinalScore,
			FinalRating:  a.FinalRating,
			Status:       a.Status,
			Comments:     a.Comments,
		})
	}
	return res, nil
}

func (s *performanceService) SubmitSelfReview(ctx context.Context, userID uint, appraisalID uint, req dto.SubmitSelfReviewRequest) error {
	appraisal, err := s.repo.FindAppraisalByID(ctx, appraisalID)
	if err != nil {
		return err
	}

	if appraisal.UserID != userID {
		return errors.New("unauthorized: this is not your appraisal")
	}

	if appraisal.Status != model.AppraisalStatusPending && appraisal.Status != model.AppraisalStatusSelfReview {
		return errors.New("cannot submit self-review at this stage")
	}

	appraisal.SelfScore = req.SelfScore
	appraisal.Comments = req.Comments
	appraisal.Status = model.AppraisalStatusManagerReview

	return s.repo.UpdateAppraisal(ctx, appraisal)
}

func (s *performanceService) SyncAttendanceKPI(ctx context.Context, userID uint, lateCount int) error {
	// Find goals with type KPI and title related to "Attendance" or "Punctuality"
	goals, err := s.repo.FindGoalsByUserID(ctx, userID)
	if err != nil {
		return err
	}

	for _, g := range goals {
		if g.Type == model.GoalTypeKPI && (g.Title == "Attendance" || g.Title == "Punctuality") {
			// Example logic: target is 0 lates, current progress is lateCount
			// Or target is 20 days, current progress is (TotalWorkDays - lateCount)
			// For now, let's just update progress as lateCount if it's a negative metric
			g.CurrentProgress = float64(lateCount)
			_ = s.repo.UpdateGoal(ctx, &g)
		}
	}
	return nil
}

func (s *performanceService) checkManagerOrAdminAccess(ctx context.Context, requesterID, targetUserID uint) error {
	if requesterID == targetUserID {
		return nil
	}

	requester, err := s.userRepo.FindByID(ctx, requesterID, []string{"role"})
	if err != nil {
		return err
	}

	// Check Admin/HR role
	roleName := requester.Role.Name
	if roleName == "superadmin" || roleName == "admin" || roleName == "hr" {
		return nil
	}

	// Check if manager
	targetUser, err := s.userRepo.FindByID(ctx, targetUserID, nil)
	if err != nil {
		return err
	}

	if targetUser.ManagerID != nil && *targetUser.ManagerID == requesterID {
		return nil
	}

	return errors.New("unauthorized: you do not have access to this user's performance data")
}

func (s *performanceService) mapGoalToResponse(g model.PerformanceGoal) dto.GoalResponse {
	return dto.GoalResponse{
		ID:              g.ID,
		UserID:          g.UserID,
		Title:           g.Title,
		Description:     g.Description,
		Type:            g.Type,
		TargetValue:     g.TargetValue,
		CurrentProgress: g.CurrentProgress,
		Unit:            g.Unit,
		StartDate:       g.StartDate.Format("2006-01-02"),
		EndDate:         g.EndDate.Format("2006-01-02"),
		Status:          g.Status,
	}
}
