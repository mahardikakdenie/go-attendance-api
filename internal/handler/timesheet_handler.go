package handler

import (
	"strconv"
	"time"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type TimesheetHandler interface {
	// Project
	CreateProject(c *gin.Context)
	GetProjects(c *gin.Context)
	UpdateProject(c *gin.Context)

	// Task
	CreateTask(c *gin.Context)
	GetTasks(c *gin.Context)

	// Timesheet
	CreateEntry(c *gin.Context)
	GetMyReport(c *gin.Context)
	GetEmployeeReport(c *gin.Context)
}

type timesheetHandler struct {
	service service.TimesheetService
}

func NewTimesheetHandler(service service.TimesheetService) TimesheetHandler {
	return &timesheetHandler{service: service}
}

// Projects
func (h *timesheetHandler) CreateProject(c *gin.Context) {
	var req model.Project
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	tenantID := c.MustGet("tenant_id").(uint)
	res, err := h.service.CreateProject(c.Request.Context(), tenantID, req)
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to create project", 500, "error", err.Error()))
		return
	}

	c.JSON(201, utils.BuildResponse("Project created successfully", 201, "success", res))
}

func (h *timesheetHandler) GetProjects(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)
	res, err := h.service.GetProjects(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to fetch projects", 500, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("Projects fetched successfully", 200, "success", res))
}

func (h *timesheetHandler) UpdateProject(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req model.Project
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	tenantID := c.MustGet("tenant_id").(uint)
	res, err := h.service.UpdateProject(c.Request.Context(), uint(id), tenantID, req)
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to update project", 500, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("Project updated successfully", 200, "success", res))
}

// Tasks
func (h *timesheetHandler) CreateTask(c *gin.Context) {
	var req model.Task
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	userID := c.MustGet("user_id").(uint)
	res, err := h.service.CreateTask(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to create task", 500, "error", err.Error()))
		return
	}

	c.JSON(201, utils.BuildResponse("Task created successfully", 201, "success", res))
}

func (h *timesheetHandler) GetTasks(c *gin.Context) {
	projectID, _ := strconv.Atoi(c.Query("project_id"))
	if projectID == 0 {
		c.JSON(400, utils.BuildErrorResponse("Project ID is required", 400, "error", nil))
		return
	}

	res, err := h.service.GetTasksByProject(c.Request.Context(), uint(projectID))
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to fetch tasks", 500, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("Tasks fetched successfully", 200, "success", res))
}

// Timesheets
func (h *timesheetHandler) CreateEntry(c *gin.Context) {
	var req model.TimesheetEntry
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	userID := c.MustGet("user_id").(uint)
	tenantID := c.MustGet("tenant_id").(uint)
	res, err := h.service.CreateTimesheet(c.Request.Context(), userID, tenantID, req)
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to create timesheet entry", 500, "error", err.Error()))
		return
	}

	c.JSON(201, utils.BuildResponse("Timesheet entry recorded", 201, "success", res))
}

func (h *timesheetHandler) GetMyReport(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))

	res, err := h.service.GetMyTimesheet(c.Request.Context(), userID, month, year)
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to generate report", 500, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("Timesheet report generated", 200, "success", res))
}

func (h *timesheetHandler) GetEmployeeReport(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("user_id"))
	month, _ := strconv.Atoi(c.DefaultQuery("month", strconv.Itoa(int(time.Now().Month()))))
	year, _ := strconv.Atoi(c.DefaultQuery("year", strconv.Itoa(time.Now().Year())))
	tenantID := c.MustGet("tenant_id").(uint)

	res, err := h.service.GetMonthlyReport(c.Request.Context(), tenantID, uint(userID), month, year)
	if err != nil {
		c.JSON(500, utils.BuildErrorResponse("Failed to generate report", 500, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("Employee timesheet report generated", 200, "success", res))
}
