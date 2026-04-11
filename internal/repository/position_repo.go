package repository

import (
	"context"
	"go-attendance-api/internal/model"
	"gorm.io/gorm"
)

type PositionRepository interface {
	Create(ctx context.Context, p *model.Position) error
	FindAll(ctx context.Context, tenantID uint) ([]model.Position, error)
	FindByID(ctx context.Context, id uint) (*model.Position, error)
	Delete(ctx context.Context, id uint) error
}

type positionRepository struct {
	db *gorm.DB
}

func NewPositionRepository(db *gorm.DB) PositionRepository {
	return &positionRepository{db: db}
}

func (r *positionRepository) Create(ctx context.Context, p *model.Position) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *positionRepository) FindAll(ctx context.Context, tenantID uint) ([]model.Position, error) {
	var results []model.Position
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("level ASC").Find(&results).Error
	return results, err
}

func (r *positionRepository) FindByID(ctx context.Context, id uint) (*model.Position, error) {
	var result model.Position
	err := r.db.WithContext(ctx).First(&result, id).Error
	return &result, err
}

func (r *positionRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.Position{}, id).Error
}
