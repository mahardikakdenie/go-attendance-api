package routes

import (
	"go-attendance-api/internal/handler"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(rg *gin.RouterGroup, h handler.AuthHandler) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", h.Register)
		auth.POST("/login", h.Login)
		auth.POST("/forgot-password", h.ForgotPassword)
		auth.POST("/reset-password", h.ResetPassword)
	}
}

func RegisterAuthenticatedAuthRoutes(rg *gin.RouterGroup, h handler.AuthHandler) {
	auth := rg.Group("/auth")
	{
		auth.GET("/sessions", h.GetSessions)
		auth.POST("/logout", h.Logout)
		auth.POST("/change-password", h.ChangePassword)
	}
}
