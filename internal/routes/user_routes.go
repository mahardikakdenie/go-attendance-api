package routes

import (
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(rg *gin.RouterGroup, userH handler.UserHandler, ucrH handler.UserChangeRequestHandler, mediaH *handler.MediaHandler) {
	users := rg.Group("/users")
	{
		users.GET("", userH.GetAllUsers)
		users.GET("/me", userH.GetMe)
		users.GET("/me/activities", userH.GetRecentActivities)
		users.POST("", middleware.RequireRole("superadmin", "admin"), userH.CreateUser)
		users.PUT("/profile-photo", userH.UpdateProfilePhoto)
		users.POST("/request-change", ucrH.CreateRequest)

		adminOnly := users.Group("")
		adminOnly.Use(middleware.RequireRole("admin", "hr"))
		{
			adminOnly.GET("/pending-changes", ucrH.GetPendingRequests)
			adminOnly.POST("/approve-change/:id", ucrH.ApproveRequest)
			adminOnly.POST("/reject-change/:id", ucrH.RejectRequest)
		}
	}

	rg.GET("/employees", userH.GetAllUsers)
	rg.POST("/media/upload", mediaH.Upload)
}
