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

	// User Payroll Profile
	GetPayrollProfile(c *gin.Context)
	UpdatePayrollProfile(c *gin.Context)
	GetMyPayrollProfile(c *gin.Context)
	GetMySlip(c *gin.Context)

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

	pagination := utils.Pagination{
		Total:       total,
		PerPage:     limit,
		CurrentPage: (offset / limit) + 1,
		LastPage:    int((total + int64(limit) - 1) / int64(limit)),
	}

	c.JSON(http.StatusOK, utils.BuildResponseWithPagination("Payroll list fetched successfully", 200, "success", data, pagination))
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

// @Summary Get Payroll Profile (Admin)
func (h *payrollHandler) GetPayrollProfile(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	res, err := h.service.GetUserPayrollProfile(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.BuildErrorResponse("Profile not found", 404, "error", nil))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

// @Summary Update Payroll Profile (Admin)
func (h *payrollHandler) UpdatePayrollProfile(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req service.UpdateUserPayrollProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	err := h.service.UpdateUserPayrollProfile(c.Request.Context(), uint(id), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Update failed", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Profile updated successfully"})
}

// @Summary Get My Payroll Profile (Self)
func (h *payrollHandler) GetMyPayrollProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	res, err := h.service.GetMyPayrollProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.BuildErrorResponse("Profile not found", 404, "error", nil))
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "data": res})
}

// @Summary Get My Detailed Slip
func (h *payrollHandler) GetMySlip(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	period := c.Query("period")
	if period == "" {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Period is required", 400, "error", nil))
		return
	}

	res, err := h.service.GetMySlip(c.Request.Context(), userID, period)
	if err != nil {
		c.JSON(http.StatusNotFound, utils.BuildErrorResponse("Slip not found", 404, "error", nil))
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
