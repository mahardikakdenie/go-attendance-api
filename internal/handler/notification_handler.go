package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type NotificationHandler interface {
	Stream(c *gin.Context)
	GetMyNotifications(c *gin.Context)
	MarkAsRead(c *gin.Context)
	MarkAllAsRead(c *gin.Context)
}

type notificationHandler struct {
	service service.NotificationService
}

func NewNotificationHandler(service service.NotificationService) NotificationHandler {
	return &notificationHandler{service: service}
}

func (h *notificationHandler) Stream(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	// Set SSE Headers
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("X-Accel-Buffering", "no") // 🆕 Disable Nginx/Proxy buffering

	// 🆕 Subscribe BEFORE starting the stream to not miss anything
	notifChan := h.service.SubscribeToUserNotifications(c.Request.Context(), userID)

	// 🆕 Heartbeat ticker
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	// 🆕 IMPORTANT: Send initial "connected" event and MANUALLY flush it.
	// This forces the proxy/browser to recognize the stream is OPEN immediately.
	c.Writer.WriteHeaderNow()
	fmt.Fprintf(c.Writer, "event: connected\ndata: {\"message\": \"Notification stream established\"}\n\n")
	c.Writer.Flush()

	// Start streaming loop
	c.Stream(func(w io.Writer) bool {
		select {
		case <-c.Request.Context().Done():
			return false
		case <-ticker.C:
			// Keep-alive ping
			fmt.Fprintf(w, ": ping\n\n")
			return true
		case notif, ok := <-notifChan:
			if !ok {
				return false
			}
			payload, _ := json.Marshal(notif)
			fmt.Fprintf(w, "data: %s\n\n", string(payload))
			return true
		}
	})
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
