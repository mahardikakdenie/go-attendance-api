package handler

import (
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type PayrollHandler interface {
	Calculate(c *gin.Context)
	Generate(c *gin.Context)
	GetList(c *gin.Context)
	GetSummary(c *gin.Context)
	Publish(c *gin.Context)
	GetMyPayrolls(c *gin.Context)

	// Individual Extensions
	GetBaseline(c *gin.Context)
	SyncAttendance(c *gin.Context)
	SaveIndividual(c *gin.Context)
}

type payrollHandler struct {
	service service.PayrollService
}

func NewPayrollHandler(service service.PayrollService) PayrollHandler {
	return &payrollHandler{service: service}
}

// @Summary Calculate Payroll (Stateless)
func (h *payrollHandler) Calculate(c *gin.Context) {
	var req service.PayrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	res, err := h.service.Calculate(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Calculation failed", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

// @Summary Generate Payroll for Period
func (h *payrollHandler) Generate(c *gin.Context) {
	var req struct {
		Period string `json:"period" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Period is required", 400, "error", err.Error()))
		return
	}

	tenantID := c.MustGet("tenant_id").(uint)
	err := h.service.GeneratePayroll(c.Request.Context(), tenantID, req.Period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Generation failed", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Payroll generated as Draft"})
}

// @Summary Get Payroll List
func (h *payrollHandler) GetList(c *gin.Context) {
	period := c.Query("period")
	if period == "" {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Period is required", 400, "error", nil))
		return
	}

	search := c.Query("search")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	tenantID := c.MustGet("tenant_id").(uint)

	data, total, err := h.service.GetAllPayroll(c.Request.Context(), tenantID, period, search, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Fetch failed", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"meta": gin.H{
			"total": total,
			"limit": limit,
			"offset": offset,
		},
		"data": data,
	})
}

// @Summary Get Payroll Summary Stats
func (h *payrollHandler) GetSummary(c *gin.Context) {
	period := c.Query("period")
	tenantID := c.MustGet("tenant_id").(uint)

	data, err := h.service.GetSummary(c.Request.Context(), tenantID, period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Fetch failed", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": data})
}

// @Summary Publish Payroll
func (h *payrollHandler) Publish(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	tenantID := c.MustGet("tenant_id").(uint)

	err := h.service.PublishPayroll(c.Request.Context(), tenantID, uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Publish failed", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Payroll published successfully"})
}

// @Summary Get Current User Payrolls
func (h *payrollHandler) GetMyPayrolls(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	res, err := h.service.GetMyPayrolls(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Fetch failed", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

// Individual Payroll Extensions

// @Summary Get Employee Baseline for Calculator
func (h *payrollHandler) GetBaseline(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("user_id"))
	res, err := h.service.GetEmployeeBaseline(c.Request.Context(), uint(userID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse(err.Error(), 500, "error", nil))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

// @Summary Sync Employee Attendance for Calculator
func (h *payrollHandler) SyncAttendance(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("user_id"))
	period := c.Query("period")
	if period == "" {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Period is required", 400, "error", nil))
		return
	}

	res, err := h.service.SyncEmployeeAttendance(c.Request.Context(), uint(userID), period)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse(err.Error(), 500, "error", nil))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

// @Summary Save Individual Payroll Record
func (h *payrollHandler) SaveIndividual(c *gin.Context) {
	userID, _ := strconv.Atoi(c.Param("user_id"))
	var req service.SaveIndividualPayrollRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request body", 400, "error", err.Error()))
		return
	}

	tenantID := c.MustGet("tenant_id").(uint)
	err := h.service.SaveIndividualPayroll(c.Request.Context(), tenantID, uint(userID), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse(err.Error(), 500, "error", nil))
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Individual payroll saved successfully"})
}
