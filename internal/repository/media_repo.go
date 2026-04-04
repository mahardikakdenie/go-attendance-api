package repository

import (
	"context"

	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type MediaRepository interface {
	Save(ctx context.Context, data *model.Media) error
}

type mediaRepository struct {
	db *gorm.DB
}

func NewMediaRepository(db *gorm.DB) MediaRepository {
	return &mediaRepository{db: db}
}

func (r *mediaRepository) Save(ctx context.Context, data *model.Media) error {
	return r.db.WithContext(ctx).Create(data).Error
}
