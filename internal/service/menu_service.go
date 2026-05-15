package service

import (
	"context"
	"encoding/json"
	"errors"
	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type MenuService interface {
	GetMyMenus(ctx context.Context, baseRole string, permissions []string, planFeatures []string, isRestricted bool) ([]model.MenuResponse, error)
	GetRolesMenuOverview(ctx context.Context, tenantID uint) ([]model.RoleMenuOverview, error)
	GetAllMenus(ctx context.Context) ([]model.Menu, error)
	UpdateMenu(ctx context.Context, id uint, req modelDto.UpdateMenuRequest) (*model.Menu, error)
	InvalidateMenuCache(ctx context.Context)
}

type menuService struct {
	repo     repository.MenuRepository
	roleRepo repository.RoleRepository
	subRepo  repository.SubscriptionRepository
	redis    *redis.Client
}

func NewMenuService(repo repository.MenuRepository, roleRepo repository.RoleRepository, subRepo repository.SubscriptionRepository, rdb *redis.Client) MenuService {
	return &menuService{repo: repo, roleRepo: roleRepo, subRepo: subRepo, redis: rdb}
}

func (s *menuService) UpdateMenu(ctx context.Context, id uint, req modelDto.UpdateMenuRequest) (*model.Menu, error) {
	allMenus, err := s.repo.FindAll(ctx) // Don't use getRawMenus here as we want a fresh copy from DB to avoid pointer issues with cache
	if err != nil {
		return nil, err
	}

	var menu *model.Menu
	for _, m := range allMenus {
		if m.ID == id {
			menu = &m
			break
		}
	}

	if menu == nil {
		return nil, errors.New("menu not found")
	}

	if req.Label != nil {
		menu.Label = *req.Label
	}
	if req.Icon != nil {
		menu.Icon = *req.Icon
	}
	if req.AllowedRoles != nil {
		menu.AllowedRoles = req.AllowedRoles
	}
	if req.SortOrder != nil {
		menu.SortOrder = *req.SortOrder
	}
	if req.IsSystem != nil {
		menu.IsSystem = *req.IsSystem
	}

	err = s.repo.Update(ctx, menu)
	if err != nil {
		return nil, err
	}

	s.InvalidateMenuCache(ctx)
	return menu, nil
}

func (s *menuService) getRawMenus(ctx context.Context) ([]model.Menu, error) {
	cacheKey := "cache:menus:all"

	// 1. Try fetching from Redis
	val, err := s.redis.Get(ctx, cacheKey).Result()
	if err == nil && val != "" {
		var menus []model.Menu
		if err := json.Unmarshal([]byte(val), &menus); err == nil {
			return menus, nil
		}
	}

	// 2. Fetch from DB if cache miss
	allMenus, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	// 3. Save to Redis (Cache for 24 hours, as menus rarely change)
	if payload, err := json.Marshal(allMenus); err == nil {
		s.redis.Set(ctx, cacheKey, payload, 24*time.Hour)
	}

	return allMenus, nil
}

func (s *menuService) InvalidateMenuCache(ctx context.Context) {
	s.redis.Del(ctx, "cache:menus:all")
}

func (s *menuService) GetAllMenus(ctx context.Context) ([]model.Menu, error) {
	return s.getRawMenus(ctx)
}

func (s *menuService) GetMyMenus(ctx context.Context, baseRole string, permissions []string, planFeatures []string, isRestricted bool) ([]model.MenuResponse, error) {
	allMenus, err := s.getRawMenus(ctx)
	if err != nil {
		return nil, err
	}

	return s.filterMenus(allMenus, baseRole, permissions, planFeatures, isRestricted), nil
}

func (s *menuService) GetRolesMenuOverview(ctx context.Context, tenantID uint) ([]model.RoleMenuOverview, error) {
	allMenus, err := s.getRawMenus(ctx)
	if err != nil {
		return nil, err
	}

	// 1. Get Plan Features
	var planFeatures []string
	sub, err := s.subRepo.FindByTenantID(ctx, tenantID)
	if err == nil && sub != nil && sub.Plan != nil {
		planFeatures = sub.Plan.Features
	} else if tenantID == 1 {
		planFeatures = []string{"*"}
	}

	// 2. Get All Roles for Tenant
	roles, err := s.roleRepo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return nil, err
	}

	// 3. For each role, calculate menus
	var overview []model.RoleMenuOverview
	for _, r := range roles {
		// FILTER: Hide platform system roles for normal tenants
		if tenantID != 1 {
			if r.BaseRole == model.BaseRoleSuperAdmin ||
				r.BaseRole == model.BaseRoleSupport ||
				r.BaseRole == model.BaseRoleEngineer {
				continue
			}
		}

		perms := make([]string, len(r.Permissions))
		for i, p := range r.Permissions {
			perms[i] = p.ID
		}

		roleMenus := s.filterMenus(allMenus, string(r.BaseRole), perms, planFeatures, false)
		overview = append(overview, model.RoleMenuOverview{
			RoleName: r.Name,
			BaseRole: string(r.BaseRole),
			Menus:    roleMenus,
		})
	}

	return overview, nil
}

func (s *menuService) filterMenus(allMenus []model.Menu, baseRole string, permissions []string, planFeatures []string, isRestricted bool) []model.MenuResponse {
	isSuperAdmin := baseRole == "SUPERADMIN"

	// 1. Filter menus in a flat list
	var filtered []model.Menu
	for _, m := range allMenus {
		// 🆕 RESTRICTION CHECK: If tenant is suspended/canceled
		if isRestricted {
			if isSuperAdmin {
				// Superadmin in restricted state can ONLY see Platform/System menus
				if !m.IsSystem {
					continue
				}
			} else {
				// Normal users/admins in restricted state can ONLY see Billing/Settings
				allowedRestrictedKeys := map[string]bool{
					"tenant-settings-billing": true,
					"governance-group":        true, // Group container
					"personal-group":          true, // Group container
					"support-desk":            true,
				}
				if !allowedRestrictedKeys[m.Key] {
					continue
				}
			}
		}

		// Rule A: System check
		if m.IsSystem && !isSuperAdmin {
			continue
		}

		// Rule B: Role check
		roleMatch := false
		for _, r := range m.AllowedRoles {
			if strings.ToUpper(r) == baseRole {
				roleMatch = true
				break
			}
		}
		if !roleMatch && !isSuperAdmin {
			continue
		}

		// Rule C: Module (Plan Feature) check
		if m.Module != "" && !isSuperAdmin {
			moduleAllowed := false
			for _, f := range planFeatures {
				if f == "*" || f == m.Module {
					moduleAllowed = true
					break
				}
			}
			if !moduleAllowed {
				continue
			}
		}

		// Rule D: Permission check
		if m.Permission != "" && !isSuperAdmin {
			hasPerm := false
			for _, p := range permissions {
				if p == m.Permission {
					hasPerm = true
					break
				}
			}
			if !hasPerm {
				continue
			}
		}

		filtered = append(filtered, m)
	}

	// 2. Build hierarchical tree
	return buildMenuTree(filtered, nil)
}

func buildMenuTree(menus []model.Menu, parentID *uint) []model.MenuResponse {
	var res []model.MenuResponse
	for _, m := range menus {
		// Match parentID
		match := false
		if parentID == nil && m.ParentID == nil {
			match = true
		} else if parentID != nil && m.ParentID != nil && *parentID == *m.ParentID {
			match = true
		}

		if match {
			children := buildMenuTree(menus, &m.ID)

			// If it's a group (path empty) but has no allowed children, skip the group
			if m.Path == "" && len(children) == 0 {
				continue
			}

			item := model.MenuResponse{
				ID:         m.ID,
				Key:        m.Key,
				Label:      m.Label,
				Icon:       m.Icon,
				Path:       m.Path,
				Module:     m.Module,
				Permission: m.Permission,
				Children:   children,
			}
			res = append(res, item)
		}
	}
	return res
}
