package service

import (
	"context"
	"encoding/json"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type MenuService interface {
	GetMyMenus(ctx context.Context, baseRole string, permissions []string, planFeatures []string, isRestricted bool) ([]model.MenuResponse, error)
	GetAllMenus(ctx context.Context) ([]model.Menu, error)
	InvalidateMenuCache(ctx context.Context)
}

type menuService struct {
	repo  repository.MenuRepository
	redis *redis.Client
}

func NewMenuService(repo repository.MenuRepository, rdb *redis.Client) MenuService {
	return &menuService{repo: repo, redis: rdb}
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
	return buildMenuTree(filtered, nil), nil
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
