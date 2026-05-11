package repository

import (
	"context"
	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type AllowancePresetRepository interface {
	FindAll(ctx context.Context) ([]model.AllowancePreset, error)
	Create(ctx context.Context, preset *model.AllowancePreset) error
}

type allowancePresetRepository struct {
	db *gorm.DB
}

func NewAllowancePresetRepository(db *gorm.DB) AllowancePresetRepository {
	return &allowancePresetRepository{db: db}
}

func (r *allowancePresetRepository) FindAll(ctx context.Context) ([]model.AllowancePreset, error) {
	var presets []model.AllowancePreset
	err := r.db.WithContext(ctx).Order("id ASC").Find(&presets).Error
	return presets, err
}

func (r *allowancePresetRepository) Create(ctx context.Context, preset *model.AllowancePreset) error {
	return r.db.WithContext(ctx).Create(preset).Error
}
