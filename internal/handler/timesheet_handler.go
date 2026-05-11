package handler

import (
	"math"
	"net/http"
	"strconv"
	"time"

	modelDto "go-attendance-api/internal/dto"
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
	DeleteProject(c *gin.Context)

	// Members
	AssignMembers(c *gin.Context)
	RemoveMember(c *gin.Context)
	GetMembers(c *gin.Context)

	// Task
	CreateTask(c *gin.Context)
	GetTasks(c *gin.Context)

	// Timesheet
	CreateEntry(c *gin.Context)
	GetMyReport(c *gin.Context)
	GetMyPaginatedReport(c *gin.Context)
	GetEmployeeReport(c *gin.Context)

	// HR Monitoring & Analytics
	GetMonitoring(c *gin.Context)
	GetAnalytics(c *gin.Context)
}

type timesheetHandler struct {
	service service.TimesheetService
}

func NewTimesheetHandler(service service.TimesheetService) TimesheetHandler {
	return &timesheetHandler{service: service}
}

// Projects
func (h *timesheetHandler) CreateProject(c *gin.Context) {
	var req model.ProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	tenantID := c.MustGet("tenant_id").(uint)
	res, err := h.service.CreateProject(c.Request.Context(), tenantID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to create project", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.BuildResponse("Project created successfully", 201, "success", res))
}

func (h *timesheetHandler) GetProjects(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)
	status := c.Query("status")
	search := c.Query("search")

	res, err := h.service.GetProjects(c.Request.Context(), tenantID, status, search)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch projects", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Projects fetched successfully", 200, "success", res))
}

func (h *timesheetHandler) UpdateProject(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req model.ProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	tenantID := c.MustGet("tenant_id").(uint)
	res, err := h.service.UpdateProject(c.Request.Context(), uint(id), tenantID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to update project", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Project updated successfully", 200, "success", res))
}

func (h *timesheetHandler) DeleteProject(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	tenantID := c.MustGet("tenant_id").(uint)

	if err := h.service.DeleteProject(c.Request.Context(), uint(id), tenantID); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to delete project", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Project deleted successfully", 200, "success", nil))
}

// Members
func (h *timesheetHandler) AssignMembers(c *gin.Context) {
	projectID, _ := strconv.Atoi(c.Param("id"))
	tenantID := c.MustGet("tenant_id").(uint)

	var req []model.ProjectMember
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	if err := h.service.AddMembers(c.Request.Context(), uint(projectID), tenantID, req); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to assign members", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Members assigned successfully", 200, "success", nil))
}

func (h *timesheetHandler) RemoveMember(c *gin.Context) {
	projectID, _ := strconv.Atoi(c.Param("id"))
	userID, _ := strconv.Atoi(c.Param("user_id"))
	tenantID := c.MustGet("tenant_id").(uint)

	if err := h.service.RemoveMember(c.Request.Context(), uint(projectID), tenantID, uint(userID)); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to remove member", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Member removed successfully", 200, "success", nil))
}

func (h *timesheetHandler) GetMembers(c *gin.Context) {
	projectID, _ := strconv.Atoi(c.Param("id"))
	tenantID := c.MustGet("tenant_id").(uint)

	res, err := h.service.GetMembers(c.Request.Context(), uint(projectID), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch members", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Members fetched successfully", 200, "success", res))
}

// Tasks
func (h *timesheetHandler) CreateTask(c *gin.Context) {
	var req model.Task
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	userID := c.MustGet("user_id").(uint)
	res, err := h.service.CreateTask(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to create task", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.BuildResponse("Task created successfully", 201, "success", res))
}

func (h *timesheetHandler) GetTasks(c *gin.Context) {
	projectID, _ := strconv.Atoi(c.Query("project_id"))
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Project ID is required", 400, "error", nil))
		return
	}

	res, err := h.service.GetTasksByProject(c.Request.Context(), uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch tasks", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Tasks fetched successfully", 200, "success", res))
}

// Timesheets
func (h *timesheetHandler) CreateEntry(c *gin.Context) {
	var req model.TimesheetEntry
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	userID := c.MustGet("user_id").(uint)
	tenantID := c.MustGet("tenant_id").(uint)
	res, err := h.service.CreateTimesheet(c.Request.Context(), userID, tenantID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to create timesheet entry", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.BuildResponse("Timesheet entry recorded", 201, "success", res))
}

func (h *timesheetHandler) GetMyReport(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	// Support Date Range filtering
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr != "" && endDateStr != "" {
		start, err1 := utils.ParseDateWIB(startDateStr)
		end, err2 := utils.ParseDateWIB(endDateStr)
		if err1 == nil && err2 == nil {
			res, err := h.service.GetMyTimesheetRange(c.Request.Context(), userID, start, end)
			if err != nil {
				c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to generate report", 500, "error", err.Error()))
				return
			}
			c.JSON(http.StatusOK, utils.BuildResponse("Timesheet report generated", 200, "success", res))
			return
		}
	}

	month := int(time.Now().Month())
	year := time.Now().Year()

	if period := c.Query("period"); period != "" {
		t, err := utils.ParseTimeWIB("2006-01", period)
		if err == nil {
			month = int(t.Month())
			year = t.Year()
		}
	} else {
		if m, err := strconv.Atoi(c.Query("month")); err == nil {
			month = m
		}
		if y, err := strconv.Atoi(c.Query("year")); err == nil {
			year = y
		}
	}

	res, err := h.service.GetMyTimesheet(c.Request.Context(), userID, month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to generate report", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Timesheet report generated", 200, "success", res))
}

func (h *timesheetHandler) GetEmployeeReport(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("user_id"))
	tenantID := c.MustGet("tenant_id").(uint)

	// Support Date Range filtering
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")

	if startDateStr != "" && endDateStr != "" {
		start, err1 := utils.ParseDateWIB(startDateStr)
		end, err2 := utils.ParseDateWIB(endDateStr)
		if err1 == nil && err2 == nil {
			res, err := h.service.GetMonthlyReportRange(c.Request.Context(), tenantID, uint(userID), start, end)
			if err != nil {
				c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to generate report", 500, "error", err.Error()))
				return
			}
			c.JSON(http.StatusOK, utils.BuildResponse("Employee timesheet report generated", 200, "success", res))
			return
		}
	}

	month := int(time.Now().Month())
	year := time.Now().Year()

	if period := c.Query("period"); period != "" {
		t, err := utils.ParseTimeWIB("2006-01", period)
		if err == nil {
			month = int(t.Month())
			year = t.Year()
		}
	} else {
		if m, err := strconv.Atoi(c.Query("month")); err == nil {
			month = m
		}
		if y, err := strconv.Atoi(c.Query("year")); err == nil {
			year = y
		}
	}

	res, err := h.service.GetMonthlyReport(c.Request.Context(), tenantID, uint(userID), month, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to generate report", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Employee timesheet report generated", 200, "success", res))
}

func (h *timesheetHandler) GetMonitoring(c *gin.Context) {
	var filter modelDto.TimesheetMonitoringFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid query parameters", 400, "error", err.Error()))
		return
	}

	tenantID := c.MustGet("tenant_id").(uint)
	res, total, err := h.service.GetMonitoring(c.Request.Context(), tenantID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch monitoring data", 500, "error", err.Error()))
		return
	}

	lastPage := int(math.Ceil(float64(total) / float64(filter.Limit)))

	c.JSON(http.StatusOK, gin.H{
		"meta": gin.H{
			"code":   200,
			"status": "success",
			"pagination": gin.H{
				"current_page": filter.Page,
				"last_page":    lastPage,
				"total":        total,
			},
		},
		"data": res,
	})
}

func (h *timesheetHandler) GetAnalytics(c *gin.Context) {
	var filter modelDto.TimesheetAnalyticsFilter
	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid query parameters", 400, "error", err.Error()))
		return
	}

	tenantID := c.MustGet("tenant_id").(uint)
	res, err := h.service.GetAnalytics(c.Request.Context(), tenantID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch analytics data", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"meta": gin.H{
			"code":   200,
			"status": "success",
		},
		"data": res,
	})
}

func (h *timesheetHandler) GetMyPaginatedReport(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	tenantID := c.MustGet("tenant_id").(uint)
	var filter modelDto.TimesheetMonitoringFilter

	if err := c.ShouldBindQuery(&filter); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid query parameters", 400, "error", err.Error()))
		return
	}

	res, err := h.service.GetMyPaginatedReport(c.Request.Context(), userID, tenantID, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to generate report", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Timesheet report generated", 200, "success", res))
}
