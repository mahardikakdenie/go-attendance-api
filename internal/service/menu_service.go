package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/events"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type MenuService interface {
	GetMyMenus(ctx context.Context, userID uint, roleID uint, baseRole string, planFeatures []string, isRestricted bool) ([]model.MenuResponse, error)
	GetRolesMenuOverview(ctx context.Context, tenantID uint) ([]model.RoleMenuOverview, error)
	GetAllMenus(ctx context.Context) ([]model.MenuResponse, error)
	CreateMenu(ctx context.Context, req modelDto.CreateMenuRequest) (*model.Menu, error)
	UpdateMenu(ctx context.Context, id uint, req modelDto.UpdateMenuRequest) (*model.MenuResponse, error)
	InvalidateMenuCache(ctx context.Context)
	InvalidateAllNavCaches(ctx context.Context)
}

type menuService struct {
	repo           repository.MenuRepository
	roleRepo       repository.RoleRepository
	subRepo        repository.SubscriptionRepository
	permissionRepo repository.PermissionRepository
	redis          *redis.Client
}

func NewMenuService(repo repository.MenuRepository, roleRepo repository.RoleRepository, subRepo repository.SubscriptionRepository, permissionRepo repository.PermissionRepository, rdb *redis.Client) MenuService {
	return &menuService{repo: repo, roleRepo: roleRepo, subRepo: subRepo, permissionRepo: permissionRepo, redis: rdb}
}

func (s *menuService) CreateMenu(ctx context.Context, req modelDto.CreateMenuRequest) (*model.Menu, error) {
	// 1. Validate ParentID
	if req.ParentID != nil {
		allMenus, err := s.repo.FindAll(ctx)
		if err != nil {
			return nil, err
		}
		found := false
		for _, m := range allMenus {
			if m.ID == *req.ParentID {
				found = true
				break
			}
		}
		if !found {
			return nil, errors.New("parent menu not found")
		}
	}

	// 2. Generate Key if not provided
	key := req.Key
	if key == "" {
		key = strings.ToLower(strings.ReplaceAll(req.Label, " ", "-"))
	}

	menu := &model.Menu{
		Key:       key,
		Label:     req.Label,
		Icon:      req.Icon,
		Path:      req.Path,
		SortOrder: req.SortOrder,
		IsSystem:  req.IsSystem,
		ParentID:  req.ParentID,
	}

	if req.RequiredPermission != "" {
		menu.RequiredPermission = &req.RequiredPermission
	}

	// Use repository to create with many2many roles if provided
	var err error
	if len(req.AllowedRoles) > 0 {
		err = s.repo.UpdateWithRoles(ctx, menu, req.AllowedRoles)
	} else {
		err = s.repo.Create(ctx, menu)
	}

	if err != nil {
		return nil, err
	}

	s.InvalidateMenuCache(ctx)
	s.InvalidateAllNavCaches(ctx)

	// Dispatch event for real-time update
	events.GetDispatcher().Dispatch(ctx, events.Event{
		Type: events.MenuChangedEvent,
		Data: menu,
	})

	return menu, nil
}

func (s *menuService) UpdateMenu(ctx context.Context, id uint, req modelDto.UpdateMenuRequest) (*model.MenuResponse, error) {
	allMenus, err := s.repo.FindAllWithRoles(ctx) // Preload roles for update
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
	if req.RequiredPermission != nil {
		if *req.RequiredPermission == "" {
			menu.RequiredPermission = nil
		} else {
			menu.RequiredPermission = req.RequiredPermission
		}
	}
	if req.SortOrder != nil {
		menu.SortOrder = *req.SortOrder
	}
	if req.IsSystem != nil {
		menu.IsSystem = *req.IsSystem
	}

	// Update roles if provided
	if req.AllowedRoles != nil {
		err = s.repo.UpdateWithRoles(ctx, menu, req.AllowedRoles)
	} else {
		err = s.repo.Update(ctx, menu)
	}

	if err != nil {
		return nil, err
	}

	// Re-fetch updated menu with roles
	updated, err := s.repo.FindAllWithRoles(ctx)
	if err != nil {
		// best-effort: return in-memory data
		roleIDs := make([]uint, len(menu.Roles))
		for j, r := range menu.Roles {
			roleIDs[j] = r.ID
		}
		s.InvalidateMenuCache(ctx)
		s.InvalidateAllNavCaches(ctx)
		return &model.MenuResponse{
			ID:           menu.ID,
			ParentID:     menu.ParentID,
			Key:          menu.Key,
			Label:        menu.Label,
			Icon:         menu.Icon,
			Path:         menu.Path,
			Module:       menu.Module,
			SortOrder:    menu.SortOrder,
			IsSystem:     menu.IsSystem,
			AllowedRoles: roleIDs,
		}, nil
	}

	var item *model.MenuResponse
	for _, u := range updated {
		if u.ID == id {
			roleIDs := make([]uint, len(u.Roles))
			for j, r := range u.Roles {
				roleIDs[j] = r.ID
			}
			item = &model.MenuResponse{
				ID:           u.ID,
				ParentID:     u.ParentID,
				Key:          u.Key,
				Label:        u.Label,
				Icon:         u.Icon,
				Path:         u.Path,
				Module:       u.Module,
				SortOrder:    u.SortOrder,
				IsSystem:     u.IsSystem,
				AllowedRoles: roleIDs,
			}
			break
		}
	}

	if item == nil {
		roleIDs := make([]uint, len(menu.Roles))
		for j, r := range menu.Roles {
			roleIDs[j] = r.ID
		}
		item = &model.MenuResponse{
			ID:           menu.ID,
			ParentID:     menu.ParentID,
			Key:          menu.Key,
			Label:        menu.Label,
			Icon:         menu.Icon,
			Path:         menu.Path,
			Module:       menu.Module,
			SortOrder:    menu.SortOrder,
			IsSystem:     menu.IsSystem,
			AllowedRoles: roleIDs,
		}
	}

	s.InvalidateMenuCache(ctx)
	s.InvalidateAllNavCaches(ctx)

	// Dispatch event for real-time update
	events.GetDispatcher().Dispatch(ctx, events.Event{
		Type: events.MenuChangedEvent,
		Data: menu,
	})

	return item, nil
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

	// 2. Fetch from DB if cache miss (Preload Roles many2many)
	allMenus, err := s.repo.FindAllWithRoles(ctx)

	println("ALLMENUS : ", allMenus)
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
	if s.redis != nil {
		s.redis.Del(ctx, "cache:menus:all")
	}
}

func (s *menuService) InvalidateAllNavCaches(ctx context.Context) {
	if s.redis != nil {
		// Use SCAN to find all user_nav:* keys and delete them
		var cursor uint64
		for {
			var keys []string
			var err error
			keys, cursor, err = s.redis.Scan(ctx, cursor, "user_nav:*", 100).Result()
			if err != nil {
				break
			}
			if len(keys) > 0 {
				s.redis.Del(ctx, keys...)
			}
			if cursor == 0 {
				break
			}
		}
	}
}

func (s *menuService) GetAllMenus(ctx context.Context) ([]model.MenuResponse, error) {
	allMenus, err := s.getRawMenus(ctx)
	if err != nil {
		return nil, err
	}

	res := make([]model.MenuResponse, len(allMenus))
	for i, m := range allMenus {
		roleIDs := make([]uint, len(m.Roles))
		for j, r := range m.Roles {
			roleIDs[j] = r.ID
		}
		res[i] = model.MenuResponse{
			ID:           m.ID,
			ParentID:     m.ParentID,
			Key:          m.Key,
			Label:        m.Label,
			Icon:         m.Icon,
			Path:         m.Path,
			Module:       m.Module,
			SortOrder:    m.SortOrder,
			IsSystem:     m.IsSystem,
			AllowedRoles: roleIDs,
		}
	}
	return res, nil
}

func (s *menuService) GetMyMenus(ctx context.Context, userID uint, roleID uint, baseRole string, planFeatures []string, isRestricted bool) ([]model.MenuResponse, error) {
	cacheKey := fmt.Sprintf("user_nav:%d:%t", userID, isRestricted)
	_ = roleID
	// roleID used in filterMenus for direct role visibility
	// cache key remains user-based and restriction-based


	// 1. Try fetching from Redis
	if s.redis != nil {
		val, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil && val != "" {
			var menus []model.MenuResponse
			if err := json.Unmarshal([]byte(val), &menus); err == nil {
				return menus, nil
			}
		}
	}

	// 2. Fetch raw and filter if cache miss
	allMenus, err := s.getRawMenus(ctx)
	if err != nil {
		return nil, err
	}

	filtered := s.filterMenus(allMenus, roleID, baseRole, planFeatures, isRestricted)

	// 3. Save to Redis (Cache for 1 hour, or until invalidated)
	if s.redis != nil {
		if payload, err := json.Marshal(filtered); err == nil {
			s.redis.Set(ctx, cacheKey, payload, 1*time.Hour)
		}
	}

	return filtered, nil
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

		roleMenus := s.filterMenus(allMenus, r.ID, string(r.BaseRole), planFeatures, false)
		overview = append(overview, model.RoleMenuOverview{
			RoleName: r.Name,
			BaseRole: string(r.BaseRole),
			Menus:    roleMenus,
		})
	}

	return overview, nil
}

func (s *menuService) filterMenus(allMenus []model.Menu, roleID uint, baseRole string, planFeatures []string, isRestricted bool) []model.MenuResponse {
	isSuperAdmin := strings.ToLower(baseRole) == string(model.BaseRoleSuperAdmin)

	// 1. Filter menus in a flat list
	var filtered []model.Menu
	for _, m := range allMenus {
		// RESTRICTION CHECK: If tenant is suspended/canceled
		if isRestricted {
			if isSuperAdmin {
				isPlatformRelated := m.IsSystem
				if m.Key == "platform-group" || (m.ParentID != nil && s.isPlatformChild(allMenus, m)) {
					isPlatformRelated = true
				}
				if !isPlatformRelated {
					fmt.Printf("[DEBUG MENU] Restricted superadmin, skip non-platform menu: key=%s label=%s\n", m.Key, m.Label)
					continue
				}
			} else {
				allowedRestrictedKeys := map[string]bool{
					"tenant-settings-billing": true,
					"governance-group":        true,
					"personal-group":          true,
					"support-desk":            true,
					"my-support":              true,
				}
				if !allowedRestrictedKeys[m.Key] {
					fmt.Printf("[DEBUG MENU] Restricted user, skip menu: key=%s label=%s\n", m.Key, m.Label)
					continue
				}
			}
		}

		// Rule A: System check
		if m.IsSystem && !isSuperAdmin {
			fmt.Printf("[DEBUG MENU] Rule A (system) skip: key=%s label=%s\n", m.Key, m.Label)
			continue
		}

		// Role-based visibility check (Primary)
		roleMapped := len(m.Roles) > 0
		hasRoleAccess := false
		if isSuperAdmin {
			hasRoleAccess = true
		} else if roleMapped {
			for _, r := range m.Roles {
				if r.ID == roleID {
					hasRoleAccess = true
					break
				}
			}
		}

		// If role is explicitly mapped and user has it, we show it (overriding module/permission)
		// If role is mapped but user doesn't have it, we skip it
		if roleMapped {
			if !hasRoleAccess {
				fmt.Printf("[DEBUG MENU] Role-mapped skip (no match): key=%s label=%s roleID=%d\n", m.Key, m.Label, roleID)
				continue
			}
		} else {
			// Not role-mapped: check old rules
			// Rule B: Module (Plan Feature) check
			if m.Module != "" && !isSuperAdmin {
				moduleAllowed := false
				for _, f := range planFeatures {
					if f == "*" || f == m.Module {
						moduleAllowed = true
						break
					}
				}
				// Always allow support/helpdesk menus even if module is not in plan features (e.g., during inactive subscription)
				if m.Key == "my-support" || m.Key == "support-desk" {
					moduleAllowed = true
				}
				if !moduleAllowed {
					fmt.Printf("[DEBUG MENU] Rule B (module) skip: key=%s label=%s module=%s planFeats=%v\n", m.Key, m.Label, m.Module, planFeatures)
					continue
				}
			}

			// Rule C: Fallback to old permission-based check for unmapped menus
			if m.RequiredPermission != nil && *m.RequiredPermission != "" && !isSuperAdmin {
				// No explicit role mapping and permission is set — skip for now as we transition
				// (Or you can keep existing permission check here if still needed)
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
		match := false
		if parentID == nil && m.ParentID == nil {
			match = true
		} else if parentID != nil && m.ParentID != nil && *parentID == *m.ParentID {
			match = true
		}

		if match {
			children := buildMenuTree(menus, &m.ID)

			if m.Path == "" && len(children) == 0 {
				continue
			}

			// Collect role IDs for the response
			roleIDs := make([]uint, len(m.Roles))
			for i, r := range m.Roles {
				roleIDs[i] = r.ID
			}

			item := model.MenuResponse{
				ID:           m.ID,
				ParentID:     m.ParentID,
				Key:          m.Key,
				Label:        m.Label,
				Icon:         m.Icon,
				Path:         m.Path,
				Module:       m.Module,
				SortOrder:    m.SortOrder,
				IsSystem:     m.IsSystem,
				AllowedRoles: roleIDs,
				Children:     children,
			}
			res = append(res, item)
		}
	}
	return res
}

func (s *menuService) isPlatformChild(allMenus []model.Menu, m model.Menu) bool {
	// Base case: if it has no parent, it's not a platform child (unless it's the platform-group itself handled elsewhere)
	if m.ParentID == nil {
		return false
	}

	// Find parent
	for _, p := range allMenus {
		if p.ID == *m.ParentID {
			if p.Key == "platform-group" {
				return true
			}
			// Recursive check for deeper nesting if needed
			return s.isPlatformChild(allMenus, p)
		}
	}
	return false
}
