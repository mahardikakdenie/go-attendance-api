package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type NotificationHandler interface {
	Stream(c *gin.Context)
	GetMyNotifications(c *gin.Context)
	MarkAsRead(c *gin.Context)
	MarkAllAsRead(c *gin.Context)
}

type notificationHandler struct {
	service service.NotificationService
	logger  *zap.Logger
}

func NewNotificationHandler(service service.NotificationService) NotificationHandler {
	return &notificationHandler{
		service: service,
		logger:  utils.GetLogger().Named("notification.handler"),
	}
}

// Stream handles the SSE endpoint for real-time push notifications.
// Supports Last-Event-ID for reconnection resume and sends initial unread_count on connect.
func (h *notificationHandler) Stream(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	logger := h.logger.With(zap.Uint("userID", userID))

	// SSE headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no")

	ctx := c.Request.Context()

	// Subscribe to Redis PubSub BEFORE sending any data
	ch, err := h.service.SubscribeToUserNotifications(ctx, userID)
	if err != nil {
		logger.Error("SSE subscribe failed", zap.Error(err))
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse(
			"Failed to establish notification stream", 500, "error", err.Error(),
		))
		return
	}

	c.Writer.WriteHeaderNow()

	// Last-Event-ID reconnection: replay missed notifications
	lastEventID := c.GetHeader("Last-Event-ID")
	if lastEventID != "" {
		if sinceID, err := strconv.ParseUint(lastEventID, 10, 64); err == nil {
			missed, err := h.service.GetNotificationsSince(ctx, userID, uint(sinceID))
			if err != nil {
				logger.Error("failed to fetch missed notifications", zap.Error(err))
			} else {
				for i := range missed {
					event := model.SSEEvent{
						Type:      "notification",
						Data:      &missed[i],
						EventID:   fmt.Sprintf("%d", missed[i].ID),
						Timestamp: utils.Now().Unix(),
					}
					h.writeSSE(c, event)
				}
				logger.Info("replayed missed notifications", zap.String("sinceID", lastEventID), zap.Int("count", len(missed)))
			}
		}
	}

	// Send initial unread count so FE badge is correct immediately
	count, err := h.service.GetUnreadCount(ctx, userID)
	if err != nil {
		logger.Error("failed to get initial unread count", zap.Error(err))
		count = 0
	}

	h.writeSSE(c, model.SSEEvent{
		Type:        "connected",
		UnreadCount: &count,
		Timestamp:   utils.Now().Unix(),
	})
	c.Writer.Flush()

	// Main SSE loop
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			fmt.Fprintf(w, ": ping\n\n")
			return true
		case msg, ok := <-ch:
			if !ok {
				return false
			}

			var event model.SSEEvent
			if err := json.Unmarshal([]byte(msg), &event); err != nil {
				logger.Error("SSE unmarshal failed", zap.Error(err), zap.String("raw", msg))
				return true // skip bad message, keep stream alive
			}

			h.writeSSE(c, event)
			return true
		}
	})

	logger.Info("SSE stream closed")
}

// writeSSE formats and writes a single SSE frame with proper event type and optional ID.
func (h *notificationHandler) writeSSE(c *gin.Context, event model.SSEEvent) {
	data, err := json.Marshal(event)
	if err != nil {
		h.logger.Error("SSE marshal failed", zap.Error(err))
		return
	}

	if event.EventID != "" {
		fmt.Fprintf(c.Writer, "id: %s\n", event.EventID)
	}
	fmt.Fprintf(c.Writer, "event: %s\n", event.Type)
	fmt.Fprintf(c.Writer, "data: %s\n\n", data)
	c.Writer.Flush()
}

func (h *notificationHandler) GetMyNotifications(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	res, err := h.service.GetMyNotifications(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch notifications", 500, "error", err.Error()))
		return
	}

	unreadCount, _ := h.service.GetUnreadCount(c.Request.Context(), userID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    res,
		"meta": gin.H{
			"unread_count": unreadCount,
		},
	})
}

func (h *notificationHandler) MarkAsRead(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	id, _ := strconv.Atoi(c.Param("id"))

	if err := h.service.MarkAsRead(c.Request.Context(), uint(id), userID); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to mark as read", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Notification marked as read", 200, "success", nil))
}

func (h *notificationHandler) MarkAllAsRead(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	if err := h.service.MarkAllAsRead(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to mark all as read", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("All notifications marked as read", 200, "success", nil))
}
