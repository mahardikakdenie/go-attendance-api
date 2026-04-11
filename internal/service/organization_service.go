package service

import (
	"context"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"time"
)

type OrganizationService interface {
	GetOrgTree(ctx context.Context, tenantID uint) ([]model.OrgNode, error)
	GetApprovalManager(ctx context.Context, userID uint, date time.Time) (*model.User, error)
	CreatePosition(ctx context.Context, p *model.Position) error
	GetPositions(ctx context.Context, tenantID uint) ([]model.Position, error)
}

type organizationService struct {
	userRepo     repository.UserRepository
	leaveRepo    repository.LeaveRepository
	positionRepo repository.PositionRepository
}

func NewOrganizationService(
	userRepo repository.UserRepository,
	leaveRepo repository.LeaveRepository,
	positionRepo repository.PositionRepository,
) OrganizationService {
	return &organizationService{
		userRepo:     userRepo,
		leaveRepo:    leaveRepo,
		positionRepo: positionRepo,
	}
}

func (s *organizationService) GetOrgTree(ctx context.Context, tenantID uint) ([]model.OrgNode, error) {
	// Fetch all users for the tenant
	users, _, err := s.userRepo.FindAll(ctx, model.UserFilter{TenantID: tenantID}, []string{"position"})
	if err != nil {
		return nil, err
	}

	// Build a map for easy access
	userMap := make(map[uint]*model.User)
	for i := range users {
		userMap[users[i].ID] = &users[i]
	}

	// Build the tree
	var roots []model.OrgNode
	for _, u := range users {
		if u.ManagerID == nil {
			roots = append(roots, s.buildNode(u, userMap))
		}
	}

	return roots, nil
}

func (s *organizationService) buildNode(u model.User, userMap map[uint]*model.User) model.OrgNode {
	posName := ""
	posLevel := 99
	if u.Position != nil {
		posName = u.Position.Name
		posLevel = u.Position.Level
	}

	node := model.OrgNode{
		ID:       u.ID,
		Name:     u.Name,
		Position: posName,
		Level:    posLevel,
		Avatar:   u.MediaUrl,
	}

	// Find subordinates
	for _, child := range userMap {
		if child.ManagerID != nil && *child.ManagerID == u.ID {
			node.Subordinates = append(node.Subordinates, s.buildNode(*child, userMap))
		}
	}

	return node
}

func (s *organizationService) GetApprovalManager(ctx context.Context, userID uint, date time.Time) (*model.User, error) {
	user, err := s.userRepo.FindByID(ctx, userID, []string{"manager"})
	if err != nil || user == nil || user.ManagerID == nil {
		return nil, err
	}

	currManager := user.Manager

	// Escalation Logic: If manager is on leave, find their manager
	for currManager != nil {
		onLeave, _ := s.leaveRepo.CheckOnLeave(ctx, currManager.ID, date)
		if !onLeave {
			return currManager, nil
		}

		// If manager is on leave, check if they have a delegate
		if currManager.DelegateID != nil {
			delegate, _ := s.userRepo.FindByID(ctx, *currManager.DelegateID, nil)
			if delegate != nil {
				return delegate, nil
			}
		}

		// Otherwise, go up the chain
		if currManager.ManagerID != nil {
			nextManager, _ := s.userRepo.FindByID(ctx, *currManager.ManagerID, []string{"manager"})
			currManager = nextManager
		} else {
			break
		}
	}

	return nil, nil
}

func (s *organizationService) CreatePosition(ctx context.Context, p *model.Position) error {
	return s.positionRepo.Create(ctx, p)
}

func (s *organizationService) GetPositions(ctx context.Context, tenantID uint) ([]model.Position, error) {
	return s.positionRepo.FindAll(ctx, tenantID)
}
