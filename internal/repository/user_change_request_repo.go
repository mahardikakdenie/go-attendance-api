package repository

import (
	"context"
	"errors"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/utils"

	"gorm.io/gorm"
)

type UserChangeRequestRepository interface {
	Create(ctx context.Context, request *model.UserChangeRequest) error
	FindByID(ctx context.Context, id uint, includes []string) (*model.UserChangeRequest, error)
	FindAll(ctx context.Context, tenantID uint, status string) ([]model.UserChangeRequest, error)
	Update(ctx context.Context, request *model.UserChangeRequest) error
}

type userChangeRequestRepository struct {
	db *gorm.DB
}

func NewUserChangeRequestRepository(db *gorm.DB) UserChangeRequestRepository {
	return &userChangeRequestRepository{
		db: db,
	}
}

var ucrPreloadMap = map[string]string{
	"user": "User",
}

func (r *userChangeRequestRepository) Create(ctx context.Context, request *model.UserChangeRequest) error {
	return r.db.WithContext(ctx).Create(request).Error
}

func (r *userChangeRequestRepository) FindByID(ctx context.Context, id uint, includes []string) (*model.UserChangeRequest, error) {
	var request model.UserChangeRequest
	query := r.db.WithContext(ctx)
	query = utils.ApplyPreloads(query, includes, ucrPreloadMap)

	err := query.First(&request, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("request not found")
		}
		return nil, err
	}
	return &request, nil
}

func (r *userChangeRequestRepository) FindAll(ctx context.Context, tenantID uint, status string) ([]model.UserChangeRequest, error) {
	var requests []model.UserChangeRequest
	query := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID)

	if status != "" {
		query = query.Where("status = ?", status)
	}

	err := query.Preload("User").Find(&requests).Error
	return requests, err
}

func (r *userChangeRequestRepository) Update(ctx context.Context, request *model.UserChangeRequest) error {
	return r.db.WithContext(ctx).Save(request).Error
}
