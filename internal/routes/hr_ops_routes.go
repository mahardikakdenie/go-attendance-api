package routes

import (
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"go-attendance-api/internal/model"
	"github.com/gin-gonic/gin"
)

func RegisterHrOpsRoutes(rg *gin.RouterGroup, h handler.HrOpsHandler) {
	hrOps := rg.Group("/hr")
	{
		hrOps.GET("/shifts", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), h.GetAllShifts)
		hrOps.POST("/shifts", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), h.CreateShift)
		hrOps.GET("/roster", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), h.GetWeeklyRoster)
		hrOps.POST("/roster/save", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), h.SaveRoster)
		hrOps.GET("/calendar", h.GetHolidays)
		hrOps.POST("/calendar", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), h.CreateHoliday)
		hrOps.PUT("/calendar/:id", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), h.UpdateHoliday)
		hrOps.DELETE("/calendar/:id", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), h.DeleteHoliday)
		hrOps.GET("/employees/:id/lifecycle", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), h.GetEmployeeLifecycle)
		hrOps.PATCH("/employees/:id/lifecycle/tasks/:task_id", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), h.UpdateLifecycleTask)
		hrOps.GET("/lifecycle-templates", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), h.GetLifecycleTemplates)
		hrOps.POST("/lifecycle-templates", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), h.CreateLifecycleTemplate)
		hrOps.DELETE("/lifecycle-templates/:id", middleware.RequireBaseRole(model.BaseRoleSuperAdmin, model.BaseRoleAdmin, model.BaseRoleHR), h.DeleteLifecycleTemplate)
	}
}
