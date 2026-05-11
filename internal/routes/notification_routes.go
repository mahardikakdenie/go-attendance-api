package routes

import (
	"go-attendance-api/internal/handler"

	"github.com/gin-gonic/gin"
)

func RegisterNotificationRoutes(rg *gin.RouterGroup, h handler.NotificationHandler) {
	notifs := rg.Group("/notifications")
	{
		notifs.GET("/stream", h.Stream)
		notifs.GET("", h.GetMyNotifications)
		notifs.PATCH("/:id/read", h.MarkAsRead)
		notifs.PATCH("/read-all", h.MarkAllAsRead)
	}
}
