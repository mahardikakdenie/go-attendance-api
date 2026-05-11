package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"

	"github.com/redis/go-redis/v9"
)

type NotificationService interface {
	SendNotification(ctx context.Context, tenantID, userID uint, title, message string, notifType model.NotificationType) error
	GetMyNotifications(ctx context.Context, userID uint, limit int) ([]model.NotificationResponse, error)
	GetUnreadCount(ctx context.Context, userID uint) (int, error)
	MarkAsRead(ctx context.Context, id uint, userID uint) error
	MarkAllAsRead(ctx context.Context, userID uint) error

	// SSE Support
	SubscribeToUserNotifications(ctx context.Context, userID uint) <-chan model.NotificationResponse
}

type notificationService struct {
	repo  repository.NotificationRepository
	redis *redis.Client
}

func NewNotificationService(repo repository.NotificationRepository, redis *redis.Client) NotificationService {
	return &notificationService{
		repo:  repo,
		redis: redis,
	}
}

func (s *notificationService) SendNotification(ctx context.Context, tenantID, userID uint, title, message string, notifType model.NotificationType) error {
	notif := &model.Notification{
		TenantID: tenantID,
		UserID:   userID,
		Title:    title,
		Message:  message,
		Type:     notifType,
		IsRead:   false,
	}

	// 1. Save to Database
	if err := s.repo.Create(ctx, notif); err != nil {
		return err
	}

	// 2. Broadcast to Redis for SSE
	channel := fmt.Sprintf("user:%d:notifications", userID)
	resp := model.NotificationResponse{
		ID:        notif.ID,
		Title:     notif.Title,
		Message:   notif.Message,
		Type:      notif.Type,
		IsRead:    notif.IsRead,
		CreatedAt: notif.CreatedAt,
	}

	payload, _ := json.Marshal(resp)
	return s.redis.Publish(ctx, channel, payload).Err()
}

func (s *notificationService) GetMyNotifications(ctx context.Context, userID uint, limit int) ([]model.NotificationResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	notifs, err := s.repo.FindAllByUser(ctx, userID, limit)
	if err != nil {
		return nil, err
	}

	var res []model.NotificationResponse
	for _, n := range notifs {
		res = append(res, model.NotificationResponse{
			ID:        n.ID,
			Title:     n.Title,
			Message:   n.Message,
			Type:      n.Type,
			IsRead:    n.IsRead,
			CreatedAt: n.CreatedAt,
		})
	}
	return res, nil
}

func (s *notificationService) GetUnreadCount(ctx context.Context, userID uint) (int, error) {
	unread, err := s.repo.FindUnreadByUser(ctx, userID)
	if err != nil {
		return 0, err
	}
	return len(unread), nil
}

func (s *notificationService) MarkAsRead(ctx context.Context, id uint, userID uint) error {
	return s.repo.MarkAsRead(ctx, id, userID)
}

func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uint) error {
	return s.repo.MarkAllAsRead(ctx, userID)
}

func (s *notificationService) SubscribeToUserNotifications(ctx context.Context, userID uint) <-chan model.NotificationResponse {
	out := make(chan model.NotificationResponse)
	channel := fmt.Sprintf("user:%d:notifications", userID)

	pubsub := s.redis.Subscribe(ctx, channel)

	go func() {
		defer pubsub.Close()
		defer close(out)

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				var notif model.NotificationResponse
				if err := json.Unmarshal([]byte(msg.Payload), &notif); err == nil {
					out <- notif
				}
			}
		}
	}()

	return out
}
