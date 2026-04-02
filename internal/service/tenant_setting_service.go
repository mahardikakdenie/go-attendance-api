package service

import (
	"context"
	"errors"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
)

type TenantSettingService interface {
	GetByTenant(ctx context.Context, tenantID uint) (*model.TenantSetting, error)
	UpdateSetting(ctx context.Context, tenantID uint, req model.TenantSetting) (*model.TenantSetting, error)
	CreateSetting(ctx context.Context, req model.TenantSetting) (*model.TenantSetting, error)
}

type tenantSettingService struct {
	repo repository.TenantSettingRepository
}

func NewTenantSettingService(repo repository.TenantSettingRepository) TenantSettingService {
	return &tenantSettingService{repo: repo}
}

func (s *tenantSettingService) GetByTenant(ctx context.Context, tenantID uint) (*model.TenantSetting, error) {
	if tenantID == 0 {
		return nil, errors.New("invalid tenant id")
	}

	return s.repo.FindByTenantID(ctx, tenantID)
}

func (s *tenantSettingService) CreateSetting(ctx context.Context, req model.TenantSetting) (*model.TenantSetting, error) {
	if req.TenantID == 0 {
		return nil, errors.New("tenant_id is required")
	}

	err := s.repo.Create(ctx, &req)
	if err != nil {
		return nil, err
	}

	return &req, nil
}

func (s *tenantSettingService) UpdateSetting(ctx context.Context, tenantID uint, req model.TenantSetting) (*model.TenantSetting, error) {
	setting, err := s.repo.FindByTenantID(ctx, tenantID)
	if err != nil {
		return nil, errors.New("tenant setting not found")
	}

	setting.OfficeLatitude = req.OfficeLatitude
	setting.OfficeLongitude = req.OfficeLongitude
	setting.MaxRadiusMeter = req.MaxRadiusMeter
	setting.AllowRemote = req.AllowRemote
	setting.RequireLocation = req.RequireLocation

	setting.ClockInStartTime = req.ClockInStartTime
	setting.ClockInEndTime = req.ClockInEndTime
	setting.LateAfterMinute = req.LateAfterMinute

	setting.ClockOutStartTime = req.ClockOutStartTime
	setting.ClockOutEndTime = req.ClockOutEndTime

	setting.RequireSelfie = req.RequireSelfie
	setting.AllowMultipleCheck = req.AllowMultipleCheck

	err = s.repo.Update(ctx, setting)
	if err != nil {
		return nil, err
	}

	return setting, nil
}
