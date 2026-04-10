package handler

import (
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PayrollHandler interface {
	Calculate(c *gin.Context)
}

type payrollHandler struct {
	service service.PayrollService
}

func NewPayrollHandler(service service.PayrollService) PayrollHandler {
	return &payrollHandler{service: service}
}

// @Summary Calculate Payroll
// @Description Dynamically calculate payroll based on TER PPh 21 & BPJS
// @Tags Payroll
// @Accept json
// @Produce json
// @Param request body service.PayrollRequest true "Payroll Data"
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} service.PayrollResponse
// @Failure 400 {object} utils.APIResponse
// @Router /api/v1/payroll/calculate [post]
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

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    res,
	})
}
