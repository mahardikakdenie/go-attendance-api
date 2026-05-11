package repository

import (
	"context"
	"go-attendance-api/internal/model"

	"gorm.io/gorm"
)

type NotificationRepository interface {
	Create(ctx context.Context, notification *model.Notification) error
	FindAllByUser(ctx context.Context, userID uint, limit int) ([]model.Notification, error)
	FindUnreadByUser(ctx context.Context, userID uint) ([]model.Notification, error)
	MarkAsRead(ctx context.Context, id uint, userID uint) error
	MarkAllAsRead(ctx context.Context, userID uint) error
}

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) Create(ctx context.Context, n *model.Notification) error {
	return r.db.WithContext(ctx).Create(n).Error
}

func (r *notificationRepository) FindAllByUser(ctx context.Context, userID uint, limit int) ([]model.Notification, error) {
	var notifications []model.Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) FindUnreadByUser(ctx context.Context, userID uint) ([]model.Notification, error) {
	var notifications []model.Notification
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND is_read = ?", userID, false).
		Order("created_at DESC").
		Find(&notifications).Error
	return notifications, err
}

func (r *notificationRepository) MarkAsRead(ctx context.Context, id uint, userID uint) error {
	return r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("id = ? AND user_id = ?", id, userID).
		Update("is_read", true).Error
}

func (r *notificationRepository) MarkAllAsRead(ctx context.Context, userID uint) error {
	return r.db.WithContext(ctx).
		Model(&model.Notification{}).
		Where("user_id = ? AND is_read = ?", userID, false).
		Update("is_read", true).Error
}
