package handler

import (
	"net/http"
	"strconv"

	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type HrOpsHandler interface {
	// Shifts
	GetAllShifts(c *gin.Context)
	CreateShift(c *gin.Context)

	// Roster
	GetWeeklyRoster(c *gin.Context)
	SaveRoster(c *gin.Context)

	// Calendar
	GetHolidays(c *gin.Context)
	CreateHoliday(c *gin.Context)
	UpdateHoliday(c *gin.Context)
	DeleteHoliday(c *gin.Context)

	// Lifecycle
	GetEmployeeLifecycle(c *gin.Context)
	UpdateLifecycleTask(c *gin.Context)
}

type hrOpsHandler struct {
	service service.HrOpsService
}

func NewHrOpsHandler(service service.HrOpsService) HrOpsHandler {
	return &hrOpsHandler{service: service}
}

// @Summary Get All Shifts
// @Tags HR Ops
// @Produce json
// @Security BearerAuth
// @Router /api/v1/hr/shifts [get]
func (h *hrOpsHandler) GetAllShifts(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)
	res, err := h.service.GetAllShifts(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch shifts", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Shifts fetched successfully", 200, "success", res))
}

// @Summary Create Shift
// @Tags HR Ops
// @Accept json
// @Produce json
// @Param body body modelDto.WorkShiftResponse true "Shift Payload"
// @Security BearerAuth
// @Router /api/v1/hr/shifts [post]
func (h *hrOpsHandler) CreateShift(c *gin.Context) {
	var req modelDto.WorkShiftResponse
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	tenantID := c.MustGet("tenant_id").(uint)
	res, err := h.service.CreateShift(c.Request.Context(), tenantID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to create shift", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, utils.BuildResponse("Shift created successfully", 201, "success", res))
}

// @Summary Get Weekly Roster
// @Tags HR Ops
// @Produce json
// @Param start_date query string true "YYYY-MM-DD"
// @Param end_date query string true "YYYY-MM-DD"
// @Param department_id query int false "Dept ID"
// @Security BearerAuth
// @Router /api/v1/hr/roster [get]
func (h *hrOpsHandler) GetWeeklyRoster(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	res, err := h.service.GetWeeklyRoster(c.Request.Context(), tenantID, startDate, endDate, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch roster", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Roster fetched successfully", 200, "success", res))
}

// @Summary Save Roster
// @Tags HR Ops
// @Accept json
// @Produce json
// @Param body body modelDto.SaveRosterRequest true "Roster Payload"
// @Security BearerAuth
// @Router /api/v1/hr/roster/save [post]
func (h *hrOpsHandler) SaveRoster(c *gin.Context) {
	var req modelDto.SaveRosterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	tenantID := c.MustGet("tenant_id").(uint)
	if err := h.service.SaveRoster(c.Request.Context(), tenantID, req); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to save roster", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Roster saved successfully", 200, "success", nil))
}

// @Summary Get Calendar/Holidays
// @Tags HR Ops
// @Produce json
// @Param year query int false "Year"
// @Security BearerAuth
// @Router /api/v1/hr/calendar [get]
func (h *hrOpsHandler) GetHolidays(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)
	yearStr := c.Query("year")
	year, _ := strconv.Atoi(yearStr)

	res, err := h.service.GetHolidays(c.Request.Context(), tenantID, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch calendar", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Calendar fetched successfully", 200, "success", res))
}

// @Summary Create Holiday
// @Tags HR Ops
// @Accept json
// @Produce json
// @Param body body modelDto.CreateHolidayRequest true "Holiday Payload"
// @Security BearerAuth
// @Router /api/v1/hr/calendar [post]
func (h *hrOpsHandler) CreateHoliday(c *gin.Context) {
	var req modelDto.CreateHolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	tenantID := c.MustGet("tenant_id").(uint)
	res, err := h.service.CreateHoliday(c.Request.Context(), tenantID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to create holiday", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusCreated, utils.BuildResponse("Holiday created successfully", 201, "success", res))
}

// @Summary Update Holiday
// @Tags HR Ops
// @Accept json
// @Produce json
// @Param id path string true "Holiday UUID"
// @Param body body modelDto.UpdateHolidayRequest true "Holiday Payload"
// @Security BearerAuth
// @Router /api/v1/hr/calendar/{id} [put]
func (h *hrOpsHandler) UpdateHoliday(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid UUID", 400, "error", err.Error()))
		return
	}

	var req modelDto.UpdateHolidayRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	tenantID := c.MustGet("tenant_id").(uint)
	if err := h.service.UpdateHoliday(c.Request.Context(), tenantID, id, req); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to update holiday", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Holiday updated successfully", 200, "success", nil))
}

// @Summary Delete Holiday
// @Tags HR Ops
// @Param id path string true "Holiday UUID"
// @Security BearerAuth
// @Router /api/v1/hr/calendar/{id} [delete]
func (h *hrOpsHandler) DeleteHoliday(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid UUID", 400, "error", err.Error()))
		return
	}

	tenantID := c.MustGet("tenant_id").(uint)
	if err := h.service.DeleteHoliday(c.Request.Context(), tenantID, id); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to delete holiday", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Holiday deleted successfully", 200, "success", nil))
}

// @Summary Get Employee Lifecycle
// @Tags HR Ops
// @Param id path int true "User ID"
// @Security BearerAuth
// @Router /api/v1/hr/employees/{id}/lifecycle [get]
func (h *hrOpsHandler) GetEmployeeLifecycle(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("id"))
	res, err := h.service.GetEmployeeLifecycle(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch lifecycle", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Lifecycle fetched successfully", 200, "success", res))
}

// @Summary Update Lifecycle Task
// @Tags HR Ops
// @Param id path int true "User ID"
// @Param task_id path string true "Task UUID"
// @Accept json
// @Param body body modelDto.UpdateLifecycleTaskRequest true "Payload"
// @Security BearerAuth
// @Router /api/v1/hr/employees/{id}/lifecycle/tasks/{task_id} [patch]
func (h *hrOpsHandler) UpdateLifecycleTask(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("id"))
	taskIDStr := c.Param("task_id")
	taskID, _ := uuid.Parse(taskIDStr)

	var req modelDto.UpdateLifecycleTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	if err := h.service.UpdateLifecycleTask(c.Request.Context(), uint(userID), taskID, req.IsCompleted); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to update task", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Task updated successfully", 200, "success", nil))
}
