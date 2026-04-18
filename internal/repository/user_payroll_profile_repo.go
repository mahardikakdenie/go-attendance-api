package repository

import (
	"context"
	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type UserPayrollProfileRepository interface {
	FindByUserID(ctx context.Context, userID uint) (*model.UserPayrollProfile, error)
	Upsert(ctx context.Context, profile *model.UserPayrollProfile) error
}

type userPayrollProfileRepository struct {
	db *gorm.DB
}

func NewUserPayrollProfileRepository(db *gorm.DB) UserPayrollProfileRepository {
	return &userPayrollProfileRepository{db: db}
}

func (r *userPayrollProfileRepository) FindByUserID(ctx context.Context, userID uint) (*model.UserPayrollProfile, error) {
	var profile model.UserPayrollProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&profile).Error
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *userPayrollProfileRepository) Upsert(ctx context.Context, profile *model.UserPayrollProfile) error {
	var existing model.UserPayrollProfile
	err := r.db.WithContext(ctx).Where("user_id = ?", profile.UserID).First(&existing).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return r.db.WithContext(ctx).Create(profile).Error
		}
		return err
	}
	
	profile.ID = existing.ID
	return r.db.WithContext(ctx).Save(profile).Error
}
