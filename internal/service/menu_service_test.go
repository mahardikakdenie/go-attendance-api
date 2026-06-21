package service

import (
	"context"
	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockMenuRepository
type MockMenuRepository struct {
	mock.Mock
}

func (m *MockMenuRepository) FindAll(ctx context.Context) ([]model.Menu, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Menu), args.Error(1)
}

func (m *MockMenuRepository) Create(ctx context.Context, menu *model.Menu) error {
	args := m.Called(ctx, menu)
	return args.Error(0)
}

func (m *MockMenuRepository) Update(ctx context.Context, menu *model.Menu) error {
	args := m.Called(ctx, menu)
	return args.Error(0)
}

func (m *MockMenuRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockMenuRepository) FindAllWithRoles(ctx context.Context) ([]model.Menu, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Menu), args.Error(1)
}

func (m *MockMenuRepository) UpdateWithRoles(ctx context.Context, menu *model.Menu, roleIDs []uint) error {
	args := m.Called(ctx, menu, roleIDs)
	return args.Error(0)
}

// MockPermissionRepository
type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) FindAll(ctx context.Context) ([]model.Permission, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Permission), args.Error(1)
}

func (m *MockPermissionRepository) FindByIDs(ctx context.Context, ids []string) ([]model.Permission, error) {
	args := m.Called(ctx, ids)
	return args.Get(0).([]model.Permission), args.Error(1)
}

// MockRoleRepository
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) FindByName(ctx context.Context, name string) (*model.Role, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRoleRepository) FindByID(ctx context.Context, id uint) (*model.Role, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Role), args.Error(1)
}

func (m *MockRoleRepository) FindByTenantID(ctx context.Context, tenantID uint) ([]model.Role, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *MockRoleRepository) FindSystemRoles(ctx context.Context) ([]model.Role, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.Role), args.Error(1)
}

func (m *MockRoleRepository) CheckRoleInUse(ctx context.Context, roleID uint) (bool, error) {
	args := m.Called(ctx, roleID)
	return args.Bool(0), args.Error(1)
}

func (m *MockRoleRepository) Create(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) Update(ctx context.Context, role *model.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) Delete(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockRoleRepository) UpdatePermissions(ctx context.Context, roleID uint, permissionIDs []string) error {
	args := m.Called(ctx, roleID, permissionIDs)
	return args.Error(0)
}

func (m *MockRoleRepository) CreateWithPermissions(ctx context.Context, role *model.Role, permissionIDs []string) error {
	args := m.Called(ctx, role, permissionIDs)
	return args.Error(0)
}

func (m *MockRoleRepository) UpdateWithPermissions(ctx context.Context, role *model.Role, permissionIDs []string) error {
	args := m.Called(ctx, role, permissionIDs)
	return args.Error(0)
}

// MockSubscriptionRepository
type MockSubscriptionRepository struct {
	mock.Mock
}

func (m *MockSubscriptionRepository) FindAll(ctx context.Context, page, limit int, status, search string) ([]model.Subscription, int64, error) {
	args := m.Called(ctx, page, limit, status, search)
	return args.Get(0).([]model.Subscription), args.Get(1).(int64), args.Error(2)
}

func (m *MockSubscriptionRepository) FindByID(ctx context.Context, id uint) (*model.Subscription, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) Create(ctx context.Context, sub *model.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) Update(ctx context.Context, sub *model.Subscription) error {
	args := m.Called(ctx, sub)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) GetStats(ctx context.Context) (float64, int64, float64, error) {
	args := m.Called(ctx)
	return args.Get(0).(float64), args.Get(1).(int64), args.Get(2).(float64), args.Error(3)
}

func (m *MockSubscriptionRepository) CountEmployees(ctx context.Context, tenantID uint) (int64, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockSubscriptionRepository) FindByTenantID(ctx context.Context, tenantID uint) (*model.Subscription, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(*model.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindPlanByName(ctx context.Context, name string) (*model.SubscriptionPlan, error) {
	args := m.Called(ctx, name)
	return args.Get(0).(*model.SubscriptionPlan), args.Error(1)
}

func (m *MockSubscriptionRepository) FindExpiringSubscriptions(ctx context.Context, days int) ([]model.Subscription, error) {
	args := m.Called(ctx, days)
	return args.Get(0).([]model.Subscription), args.Error(1)
}

func (m *MockSubscriptionRepository) FindAllPlans(ctx context.Context) ([]model.SubscriptionPlan, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.SubscriptionPlan), args.Error(1)
}

func (m *MockSubscriptionRepository) FindPlanByID(ctx context.Context, id uint) (*model.SubscriptionPlan, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*model.SubscriptionPlan), args.Error(1)
}

func (m *MockSubscriptionRepository) CreatePlan(ctx context.Context, plan *model.SubscriptionPlan) error {
	args := m.Called(ctx, plan)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) UpdatePlan(ctx context.Context, plan *model.SubscriptionPlan) error {
	args := m.Called(ctx, plan)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) DeletePlan(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockSubscriptionRepository) FindAllFeatures(ctx context.Context) ([]model.SubscriptionFeature, error) {
	args := m.Called(ctx)
	return args.Get(0).([]model.SubscriptionFeature), args.Error(1)
}

func TestCreateMenu(t *testing.T) {
	repo := new(MockMenuRepository)
	roleRepo := new(MockRoleRepository)
	subRepo := new(MockSubscriptionRepository)
	permRepo := new(MockPermissionRepository)
	s := NewMenuService(repo, roleRepo, subRepo, permRepo, nil)

	ctx := context.Background()

	t.Run("Success - Create Parent Menu", func(t *testing.T) {
		req := modelDto.CreateMenuRequest{
			Label: "Test Menu",
			Icon:  "TestIcon",
			Path:  "/test",
		}

		repo.On("Create", ctx, mock.AnythingOfType("*model.Menu")).Return(nil).Once()

		res, err := s.CreateMenu(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "test-menu", res.Key)
		assert.Equal(t, "Test Menu", res.Label)
		repo.AssertExpectations(t)
	})

	t.Run("Success - Create Sub Menu with Permission", func(t *testing.T) {
		parentID := uint(1)
		req := modelDto.CreateMenuRequest{
			Label:              "Sub Menu",
			Icon:               "SubIcon",
			Path:               "/sub",
			ParentID:           &parentID,
			RequiredPermission: "test.perm",
		}

		repo.On("FindAll", ctx).Return([]model.Menu{{ID: 1}}, nil).Once()
		repo.On("Create", ctx, mock.AnythingOfType("*model.Menu")).Return(nil).Once()

		res, err := s.CreateMenu(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "sub-menu", res.Key)
		assert.Equal(t, &parentID, res.ParentID)
		assert.Equal(t, "test.perm", *res.RequiredPermission)
		permRepo.AssertExpectations(t)
		repo.AssertExpectations(t)
	})

	t.Run("Success - Create Menu with RequiredPermission", func(t *testing.T) {
		req := modelDto.CreateMenuRequest{
			Label:              "Test Menu",
			Icon:               "TestIcon",
			RequiredPermission: "invalid.perm",
		}

		repo.On("Create", ctx, mock.AnythingOfType("*model.Menu")).Return(nil).Once()

		res, err := s.CreateMenu(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, res)
		assert.Equal(t, "invalid.perm", *res.RequiredPermission)
		repo.AssertExpectations(t)
	})

	t.Run("Failure - Invalid Parent ID", func(t *testing.T) {
		parentID := uint(99)
		req := modelDto.CreateMenuRequest{
			Label:    "Test Menu",
			Icon:     "TestIcon",
			ParentID: &parentID,
		}

		repo.On("FindAll", ctx).Return([]model.Menu{{ID: 1}}, nil).Once()

		res, err := s.CreateMenu(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, res)
		assert.Equal(t, "parent menu not found", err.Error())
		repo.AssertExpectations(t)
	})
}
