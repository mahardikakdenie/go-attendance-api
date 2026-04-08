package repository

import (
	"context"
	"go-attendance-api/internal/model"
	"gorm.io/gorm"
)

type RecentActivityRepository interface {
	FindByUserID(ctx context.Context, userID uint, limit int) ([]model.RecentActivity, error)
	Create(ctx context.Context, activity *model.RecentActivity) error
}

type recentActivityRepository struct {
	db *gorm.DB
}

func NewRecentActivityRepository(db *gorm.DB) RecentActivityRepository {
	return &recentActivityRepository{
		db: db,
	}
}

func (r *recentActivityRepository) FindByUserID(ctx context.Context, userID uint, limit int) ([]model.RecentActivity, error) {
	var activities []model.RecentActivity
	
	query := r.db.WithContext(ctx).Where("user_id = ?", userID).Order("created_at DESC")
	
	if limit > 0 {
		query = query.Limit(limit)
	}
	
	err := query.Find(&activities).Error
	return activities, err
}

func (r *recentActivityRepository) Create(ctx context.Context, activity *model.RecentActivity) error {
	return r.db.WithContext(ctx).Create(activity).Error
}
