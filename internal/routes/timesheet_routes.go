package routes

import (
	"go-attendance-api/internal/handler"
	"go-attendance-api/internal/middleware"
	"github.com/gin-gonic/gin"
)

func RegisterTimesheetRoutes(rg *gin.RouterGroup, h handler.TimesheetHandler) {
	timesheet := rg.Group("/timesheet")
	{
		timesheet.POST("/entries", h.CreateEntry)
		timesheet.GET("/me/report", h.GetMyReport)
		timesheet.POST("/tasks", h.CreateTask)
		timesheet.GET("/tasks", h.GetTasks)
		
		timesheet.GET("/projects", h.GetProjects)
		timesheet.GET("/projects/:id/members", h.GetMembers)

		adminTimesheet := timesheet.Group("/admin")
		adminTimesheet.Use(middleware.RequireRole("superadmin", "admin", "hr"))
		{
			adminTimesheet.POST("/projects", h.CreateProject)
			adminTimesheet.PUT("/projects/:id", h.UpdateProject)
			adminTimesheet.DELETE("/projects/:id", h.DeleteProject)
			
			adminTimesheet.POST("/projects/:id/members", h.AssignMembers)
			adminTimesheet.DELETE("/projects/:id/members/:user_id", h.RemoveMember)

			adminTimesheet.GET("/report/employee/:user_id", h.GetEmployeeReport)
		}
	}

	projects := rg.Group("/projects")
	{
		projects.GET("", middleware.HasPermission("user.view"), h.GetProjects)
		projects.POST("", middleware.HasPermission("project.manage"), h.CreateProject)
		projects.PUT("/:id", middleware.HasPermission("project.manage"), h.UpdateProject)
		projects.DELETE("/:id", middleware.HasPermission("project.manage"), h.DeleteProject)
		
		projects.POST("/:id/members", middleware.HasPermission("project.manage"), h.AssignMembers)
		projects.DELETE("/:id/members/:user_id", middleware.HasPermission("project.manage"), h.RemoveMember)
		projects.GET("/:id/members", middleware.HasPermission("user.view"), h.GetMembers)
	}
}
