package routes

import (
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"go-attendance-api/internal/model"
	"github.com/gin-gonic/gin"
)

func RegisterDashboardRoutes(rg *gin.RouterGroup, h handler.DashboardHandler) {
	dashboards := rg.Group("/dashboards")
	{
		dashboards.GET("/admin", middleware.RequireBaseRole(model.BaseRoleSuperAdmin), h.GetAdminDashboard)
		dashboards.GET("/hr", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), h.GetHrDashboard)
		dashboards.GET("/hr/daily-pulse", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), h.GetDailyPulse)
		dashboards.GET("/hr/employee-dna/:user_id", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), h.GetEmployeeDNA)
		dashboards.GET("/finance", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleFinance), h.GetFinanceDashboard)
		dashboards.GET("/heatmap", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), h.GetHeatmap)
	}
}
