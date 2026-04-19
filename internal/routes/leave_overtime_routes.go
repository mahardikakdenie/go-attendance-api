package routes

import (
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterLeaveRoutes(rg *gin.RouterGroup, h handler.LeaveHandler) {
	leaves := rg.Group("/leaves")
	{
		leaves.POST("/request", h.RequestLeave)
		leaves.GET("", h.GetLeaveHistory)
		leaves.GET("/balances", h.GetLeaveBalances)
		leaves.POST("/approve/:id", middleware.RequireRole("superadmin", "admin", "hr"), h.ApproveLeave)
		leaves.POST("/reject/:id", middleware.RequireRole("superadmin", "admin", "hr"), h.RejectLeave)
	}
}

func RegisterOvertimeRoutes(rg *gin.RouterGroup, h handler.OvertimeHandler) {
	overtime := rg.Group("/overtime")
	{
		overtime.POST("", h.CreateRequest)
		overtime.GET("", h.GetAll)
		overtime.GET("/:id", h.GetByID)

		overtime.POST("/approve/:id", middleware.RequireRole("superadmin", "admin", "hr"), h.ApproveRequest)
		overtime.POST("/reject/:id", middleware.RequireRole("superadmin", "admin", "hr"), h.RejectRequest)
	}
}
