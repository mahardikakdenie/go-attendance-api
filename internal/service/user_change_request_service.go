package service

import (
	"context"
	"errors"
	"fmt"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/utils"
)

type UserChangeRequestService interface {
	CreateRequest(ctx context.Context, userID uint, tenantID uint, req model.CreateUserChangeRequest) (model.UserChangeRequestResponse, error)
	GetMyRequests(ctx context.Context, userID uint) ([]model.UserChangeRequestResponse, error)
	GetAllRequests(ctx context.Context, tenantID uint, status string) ([]model.UserChangeRequestResponse, error)
	ApproveRequest(ctx context.Context, requestID uint, adminID uint) error
	RejectRequest(ctx context.Context, requestID uint, adminID uint, notes string) error
	CancelRequest(ctx context.Context, requestID uint, userID uint) error
}

type userChangeRequestService struct {
	repo         repository.UserChangeRequestRepository
	userRepo     repository.UserRepository
	notifService NotificationService
}

func NewUserChangeRequestService(repo repository.UserChangeRequestRepository, userRepo repository.UserRepository, notifService NotificationService) UserChangeRequestService {
	return &userChangeRequestService{
		repo:         repo,
		userRepo:     userRepo,
		notifService: notifService,
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

	// 🆕 NOTIFICATION: Notify HR/Admin of the tenant
	// Find admins in this tenant
	admins, _, _ := s.userRepo.FindAll(ctx, model.UserFilter{TenantID: tenantID}, nil)
	for _, admin := range admins {
		// Only notify those who have permission to manage requests or are admin/hr roles
		// Simple check: if role name is admin or hr
		if admin.Role != nil && (admin.Role.Name == "admin" || admin.Role.Name == "hr") {
			user, _ := s.userRepo.FindByID(ctx, userID, nil)
			userName := "An employee"
			if user != nil {
				userName = user.Name
			}
			s.notifService.SendNotification(ctx, tenantID, admin.ID, "Profile Change Request", fmt.Sprintf("%s has requested a profile update.", userName), model.NotificationTypeProfile)
		}
	}

	return mapToUCRResponse(request), nil
}

func (s *userChangeRequestService) GetMyRequests(ctx context.Context, userID uint) ([]model.UserChangeRequestResponse, error) {
	requests, err := s.repo.FindAllByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	var responses []model.UserChangeRequestResponse
	for _, req := range requests {
		responses = append(responses, mapToUCRResponse(&req))
	}

	return responses, nil
}

func (s *userChangeRequestService) GetAllRequests(ctx context.Context, tenantID uint, status string) ([]model.UserChangeRequestResponse, error) {
	requests, err := s.repo.FindAll(ctx, tenantID, status)
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
	user.UpdatedAt = utils.Now()

	if err := s.userRepo.Update(ctx, user); err != nil {
		return err
	}

	// Update Request status
	now := utils.Now()
	request.Status = model.StatusApproved
	request.ApprovedBy = &adminID
	request.ApprovedAt = &now

	// NOTIFICATION
	s.notifService.SendNotification(ctx, request.TenantID, request.UserID, "Profile Update Approved", "Your profile update request has been approved and applied.", model.NotificationTypeProfile)

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
	now := utils.Now()
	request.ApprovedAt = &now

	// NOTIFICATION
	s.notifService.SendNotification(ctx, request.TenantID, request.UserID, "Profile Update Rejected", "Your profile update request has been rejected.", model.NotificationTypeProfile)

	return s.repo.Update(ctx, request)
}

func (s *userChangeRequestService) CancelRequest(ctx context.Context, requestID uint, userID uint) error {
	request, err := s.repo.FindByID(ctx, requestID, []string{})
	if err != nil {
		return err
	}

	if request.UserID != userID {
		return errors.New("unauthorized to cancel this request")
	}

	if request.Status != model.StatusPending && request.Status != model.StatusDraft {
		return errors.New("can only cancel pending or draft requests")
	}

	request.Status = model.StatusCancelled
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
