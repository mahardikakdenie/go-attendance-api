package routes

import (
	"go-attendance-api/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterMiscRoutes(rg *gin.RouterGroup, activityH handler.ActivityHandler) {
	activities := rg.Group("/activities")
	{
		activities.GET("/recent", activityH.GetRecentActivities)
		activities.GET("/quick-info", activityH.GetQuickInfo)
	}
}
