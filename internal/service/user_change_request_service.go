package service

import (
	"context"
	"errors"
	"time"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
)

type UserChangeRequestService interface {
	CreateRequest(ctx context.Context, userID uint, tenantID uint, req model.CreateUserChangeRequest) (model.UserChangeRequestResponse, error)
	GetPendingRequests(ctx context.Context, tenantID uint) ([]model.UserChangeRequestResponse, error)
	ApproveRequest(ctx context.Context, requestID uint, adminID uint) error
	RejectRequest(ctx context.Context, requestID uint, adminID uint, notes string) error
}

type userChangeRequestService struct {
	repo     repository.UserChangeRequestRepository
	userRepo repository.UserRepository
}

func NewUserChangeRequestService(repo repository.UserChangeRequestRepository, userRepo repository.UserRepository) UserChangeRequestService {
	return &userChangeRequestService{
		repo:     repo,
		userRepo: userRepo,
	}
}

func (s *userChangeRequestService) CreateRequest(ctx context.Context, userID uint, tenantID uint, req model.CreateUserChangeRequest) (model.UserChangeRequestResponse, error) {
	request := &model.UserChangeRequest{
		UserID:      userID,
		TenantID:    tenantID,
		Name:        req.Name,
		Email:       req.Email,
		Department:  req.Department,
		Address:     req.Address,
		PhoneNumber: req.PhoneNumber,
		Status:      model.StatusPending,
	}

	if err := s.repo.Create(ctx, request); err != nil {
		return model.UserChangeRequestResponse{}, err
	}

	return mapToUCRResponse(request), nil
}

func (s *userChangeRequestService) GetPendingRequests(ctx context.Context, tenantID uint) ([]model.UserChangeRequestResponse, error) {
	requests, err := s.repo.FindAll(ctx, tenantID, string(model.StatusPending))
	if err != nil {
		return nil, err
	}

	var responses []model.UserChangeRequestResponse
	for _, req := range requests {
		responses = append(responses, mapToUCRResponse(&req))
	}

	return responses, nil
}

func (s *userChangeRequestService) ApproveRequest(ctx context.Context, requestID uint, adminID uint) error {
	request, err := s.repo.FindByID(ctx, requestID, []string{"user"})
	if err != nil {
		return err
	}

	if request.Status != model.StatusPending {
		return errors.New("request is not pending")
	}

	// Update User data
	user, err := s.userRepo.FindByID(ctx, request.UserID, []string{})
	if err != nil {
		return err
	}

	user.Name = request.Name
	user.Email = request.Email
	user.Department = request.Department
	user.Address = request.Address
	user.PhoneNumber = request.PhoneNumber
	user.UpdatedAt = time.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Update Request status
	now := time.Now()
	request.Status = model.StatusApproved
	request.ApprovedBy = &adminID
	request.ApprovedAt = &now

	return s.repo.Update(ctx, request)
}

func (s *userChangeRequestService) RejectRequest(ctx context.Context, requestID uint, adminID uint, notes string) error {
	request, err := s.repo.FindByID(ctx, requestID, []string{})
	if err != nil {
		return err
	}

	if request.Status != model.StatusPending {
		return errors.New("request is not pending")
	}

	request.Status = model.StatusRejected
	request.AdminNotes = notes
	request.ApprovedBy = &adminID
	now := time.Now()
	request.ApprovedAt = &now

	return s.repo.Update(ctx, request)
}

func mapToUCRResponse(req *model.UserChangeRequest) model.UserChangeRequestResponse {
	res := model.UserChangeRequestResponse{
		ID:          req.ID,
		UserID:      req.UserID,
		TenantID:    req.TenantID,
		Name:        req.Name,
		Email:       req.Email,
		Department:  req.Department,
		Address:     req.Address,
		PhoneNumber: req.PhoneNumber,
		Status:      req.Status,
		AdminNotes:  req.AdminNotes,
		CreatedAt:   req.CreatedAt,
		UpdatedAt:   req.UpdatedAt,
	}

	if req.User != nil {
		var roleRes *model.RoleResponse
		if req.User.Role != nil {
			roleRes = &model.RoleResponse{
				ID:   req.User.Role.ID,
				Name: req.User.Role.Name,
			}
		}

		res.User = &model.UserResponse{
			ID:          req.User.ID,
			Name:        req.User.Name,
			Email:       req.User.Email,
			Role:        roleRes,
			TenantID:    req.User.TenantID,
			EmployeeID:  req.User.EmployeeID,
			Department:  req.User.Department,
			Address:     req.User.Address,
			PhoneNumber: req.User.PhoneNumber,
		}
	}

	return res
}
