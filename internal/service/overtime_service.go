package service

import (
	"context"
	"errors"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"time"
)

type OvertimeService interface {
	CreateRequest(ctx context.Context, userID uint, tenantID uint, req model.CreateOvertimeRequest) (model.OvertimeResponse, error)
	ApproveRequest(ctx context.Context, id uint, adminID uint, req model.ApproveOvertimeRequest) (model.OvertimeResponse, error)
	RejectRequest(ctx context.Context, id uint, adminID uint, req model.ApproveOvertimeRequest) (model.OvertimeResponse, error)
	GetAll(ctx context.Context, requesterID uint, filter model.OvertimeFilter) ([]model.OvertimeResponse, int64, error)
	GetByID(ctx context.Context, id uint) (model.OvertimeResponse, error)
	GetPendingCount(ctx context.Context, userID uint) (int, error)
}

type overtimeService struct {
	repo        repository.OvertimeRepository
	userService UserService
}

func NewOvertimeService(repo repository.OvertimeRepository, userService UserService) OvertimeService {
	return &overtimeService{
		repo:        repo,
		userService: userService,
	}
}

func (s *overtimeService) GetPendingCount(ctx context.Context, userID uint) (int, error) {
	count, err := s.repo.GetPendingCount(ctx, userID)
	return int(count), err
}

func (s *overtimeService) CreateRequest(ctx context.Context, userID uint, tenantID uint, req model.CreateOvertimeRequest) (model.OvertimeResponse, error) {
	date, err := time.Parse("2006-01-02", req.Date)
	if err != nil {
		return model.OvertimeResponse{}, errors.New("invalid date format, use YYYY-MM-DD")
	}

	overtime := model.Overtime{
		UserID:    userID,
		TenantID:  tenantID,
		Date:      date,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Reason:    req.Reason,
		Status:    model.OvertimeStatusPending,
	}

	if err := s.repo.Save(ctx, &overtime); err != nil {
		return model.OvertimeResponse{}, err
	}

	return s.mapToResponse(overtime), nil
}

func (s *overtimeService) ApproveRequest(ctx context.Context, id uint, adminID uint, req model.ApproveOvertimeRequest) (model.OvertimeResponse, error) {
	overtime, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return model.OvertimeResponse{}, errors.New("overtime request not found")
	}

	if overtime.Status != model.OvertimeStatusPending {
		return model.OvertimeResponse{}, errors.New("request is already processed")
	}

	now := time.Now()
	overtime.Status = model.OvertimeStatusApproved
	overtime.AdminNotes = req.AdminNotes
	overtime.ApprovedBy = &adminID
	overtime.ApprovedAt = &now

	if err := s.repo.Update(ctx, overtime); err != nil {
		return model.OvertimeResponse{}, err
	}

	return s.mapToResponse(*overtime), nil
}

func (s *overtimeService) RejectRequest(ctx context.Context, id uint, adminID uint, req model.ApproveOvertimeRequest) (model.OvertimeResponse, error) {
	overtime, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return model.OvertimeResponse{}, errors.New("overtime request not found")
	}

	if overtime.Status != model.OvertimeStatusPending {
		return model.OvertimeResponse{}, errors.New("request is already processed")
	}

	now := time.Now()
	overtime.Status = model.OvertimeStatusRejected
	overtime.AdminNotes = req.AdminNotes
	overtime.ApprovedBy = &adminID
	overtime.ApprovedAt = &now

	if err := s.repo.Update(ctx, overtime); err != nil {
		return model.OvertimeResponse{}, err
	}

	return s.mapToResponse(*overtime), nil
}

func (s *overtimeService) GetAll(ctx context.Context, requesterID uint, filter model.OvertimeFilter) ([]model.OvertimeResponse, int64, error) {
	// Apply Hierarchical Scoping
	if requesterID != 0 {
		allowedRoleIDs, _ := s.userService.GetAllowedRoleIDs(ctx, requesterID)
		filter.AllowedRoleIDs = allowedRoleIDs
	}

	data, total, err := s.repo.FindAll(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	var responses []model.OvertimeResponse
	for _, o := range data {
		responses = append(responses, s.mapToResponse(o))
	}

	return responses, total, nil
}

func (s *overtimeService) GetByID(ctx context.Context, id uint) (model.OvertimeResponse, error) {
	overtime, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return model.OvertimeResponse{}, err
	}
	return s.mapToResponse(*overtime), nil
}

func (s *overtimeService) mapToResponse(o model.Overtime) model.OvertimeResponse {
	res := model.OvertimeResponse{
		ID:         o.ID,
		UserID:     o.UserID,
		Date:       o.Date,
		StartTime:  o.StartTime,
		EndTime:    o.EndTime,
		Reason:     o.Reason,
		Status:     o.Status,
		AdminNotes: o.AdminNotes,
		CreatedAt:  o.CreatedAt,
		UpdatedAt:  o.UpdatedAt,
	}

	if o.User != nil {
		var roleRes *model.RoleResponse
		if o.User.Role != nil {
			roleRes = &model.RoleResponse{
				ID:   o.User.Role.ID,
				Name: o.User.Role.Name,
			}
		}

		res.User = &model.UserResponse{
			ID:         o.User.ID,
			Name:       o.User.Name,
			Email:      o.User.Email,
			Role:       roleRes,
			EmployeeID: o.User.EmployeeID,
			Department: o.User.Department,
		}
	}

	return res
}
