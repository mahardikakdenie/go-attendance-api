package service

import (
	"context"
	"encoding/json"
	"fmt"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/repository"
	"go-attendance-api/internal/utils"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// channelKey builds the Redis PubSub channel name for a user's notification stream.
func channelKey(userID uint) string {
	return fmt.Sprintf("user:%d:notifications", userID)
}

type NotificationService interface {
	SendNotification(ctx context.Context, tenantID, userID uint, title, message string, notifType model.NotificationType) error
	GetMyNotifications(ctx context.Context, userID uint, limit int) ([]model.NotificationResponse, error)
	GetUnreadCount(ctx context.Context, userID uint) (int64, error)
	GetNotificationsSince(ctx context.Context, userID uint, sinceID uint) ([]model.NotificationResponse, error)
	MarkAsRead(ctx context.Context, id uint, userID uint) error
	MarkAllAsRead(ctx context.Context, userID uint) error

	// SSE Support
	SubscribeToUserNotifications(ctx context.Context, userID uint) (<-chan string, error)
	BroadcastSSEEvent(ctx context.Context, event model.SSEEvent) error
}

type notificationService struct {
	repo   repository.NotificationRepository
	redis  *redis.Client
	logger *zap.Logger
}

func NewNotificationService(repo repository.NotificationRepository, rdb *redis.Client) NotificationService {
	return &notificationService{
		repo:   repo,
		redis:  rdb,
		logger: utils.GetLogger().Named("notification"),
	}
}

const GlobalChannel = "global:notifications"

// SendNotification saves to DB, then publishes SSEEvent with live unread_count to Redis.
func (s *notificationService) SendNotification(ctx context.Context, tenantID, userID uint, title, message string, notifType model.NotificationType) error {
	notif := &model.Notification{
		TenantID: tenantID,
		UserID:   userID,
		Title:    title,
		Message:  message,
		Type:     notifType,
		IsRead:   false,
	}

	if err := s.repo.Create(ctx, notif); err != nil {
		return err
	}

	// Fetch fresh count AFTER insert for accuracy
	count, err := s.repo.CountUnreadByUser(ctx, userID)
	if err != nil {
		s.logger.Error("failed to count unread after send", zap.Uint("userID", userID), zap.Error(err))
		count = 0
	}

	resp := &model.NotificationResponse{
		ID:        notif.ID,
		Title:     notif.Title,
		Message:   notif.Message,
		Type:      notif.Type,
		IsRead:    notif.IsRead,
		CreatedAt: notif.CreatedAt,
	}

	event := model.SSEEvent{
		Type:        "notification",
		UnreadCount: &count,
		Data:        resp,
		EventID:     fmt.Sprintf("%d", notif.ID),
		Timestamp:   utils.Now().Unix(),
	}

	return s.publishSSEEvent(ctx, userID, event)
}

// GetMyNotifications returns paginated notifications for a user.
func (s *notificationService) GetMyNotifications(ctx context.Context, userID uint, limit int) ([]model.NotificationResponse, error) {
	if limit <= 0 {
		limit = 20
	}
	notifs, err := s.repo.FindAllByUser(ctx, userID, limit)
	if err != nil {
		return nil, err
	}
	return toNotificationResponses(notifs), nil
}

// GetUnreadCount returns the count of unread notifications via SQL COUNT.
func (s *notificationService) GetUnreadCount(ctx context.Context, userID uint) (int64, error) {
	return s.repo.CountUnreadByUser(ctx, userID)
}

// GetNotificationsSince returns notifications created after a given ID (for Last-Event-ID reconnection).
func (s *notificationService) GetNotificationsSince(ctx context.Context, userID uint, sinceID uint) ([]model.NotificationResponse, error) {
	notifs, err := s.repo.FindSinceID(ctx, userID, sinceID, 50)
	if err != nil {
		return nil, err
	}
	return toNotificationResponses(notifs), nil
}

// MarkAsRead marks a single notification as read, then pushes updated count via SSE.
func (s *notificationService) MarkAsRead(ctx context.Context, id uint, userID uint) error {
	if err := s.repo.MarkAsRead(ctx, id, userID); err != nil {
		return err
	}

	go s.publishUnreadCountUpdate(userID)
	return nil
}

// MarkAllAsRead marks all notifications as read, then pushes updated count via SSE.
func (s *notificationService) MarkAllAsRead(ctx context.Context, userID uint) error {
	if err := s.repo.MarkAllAsRead(ctx, userID); err != nil {
		return err
	}

	go s.publishUnreadCountUpdate(userID)
	return nil
}

// SubscribeToUserNotifications opens a Redis PubSub subscription and returns a channel of raw JSON strings.
// It subscribes to both user-specific and global channels.
func (s *notificationService) SubscribeToUserNotifications(ctx context.Context, userID uint) (<-chan string, error) {
	pubsub := s.redis.Subscribe(ctx, channelKey(userID), GlobalChannel)

	// Confirm subscription is active before returning
	if _, err := pubsub.Receive(ctx); err != nil {
		_ = pubsub.Close()
		return nil, fmt.Errorf("redis subscribe failed: %w", err)
	}

	out := make(chan string, 16) // buffered to absorb bursts

	go func() {
		defer close(out)
		defer func() {
			if err := pubsub.Close(); err != nil {
				s.logger.Error("pubsub close error", zap.Uint("userID", userID), zap.Error(err))
			}
		}()

		ch := pubsub.Channel()
		for {
			select {
			case <-ctx.Done():
				return
			case msg, ok := <-ch:
				if !ok {
					return
				}
				// Double-select: if context cancelled while waiting to send, exit cleanly
				select {
				case out <- msg.Payload:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return out, nil
}

// publishUnreadCountUpdate pushes an "unread_count" SSE event. Fire-and-forget (called in goroutine).
func (s *notificationService) publishUnreadCountUpdate(userID uint) {
	ctx := context.Background()

	count, err := s.repo.CountUnreadByUser(ctx, userID)
	if err != nil {
		s.logger.Error("failed to count unread for SSE update", zap.Uint("userID", userID), zap.Error(err))
		return
	}

	event := model.SSEEvent{
		Type:        "unread_count",
		UnreadCount: &count,
		Timestamp:   utils.Now().Unix(),
	}

	if err := s.publishSSEEvent(ctx, userID, event); err != nil {
		s.logger.Error("failed to publish unread count event", zap.Uint("userID", userID), zap.Error(err))
	}
}

// publishSSEEvent marshals and publishes an SSEEvent to the user's Redis channel.
func (s *notificationService) publishSSEEvent(ctx context.Context, userID uint, event model.SSEEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal SSE event: %w", err)
	}
	return s.redis.Publish(ctx, channelKey(userID), payload).Err()
}

// BroadcastSSEEvent publishes an SSEEvent to the global channel.
func (s *notificationService) BroadcastSSEEvent(ctx context.Context, event model.SSEEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal SSE event for broadcast: %w", err)
	}
	return s.redis.Publish(ctx, GlobalChannel, payload).Err()
}

// toNotificationResponses maps model slices to response DTOs.
func toNotificationResponses(notifs []model.Notification) []model.NotificationResponse {
	res := make([]model.NotificationResponse, 0, len(notifs))
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
	return res
}
