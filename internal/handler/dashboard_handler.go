package handler

import (
	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DashboardHandler interface {
	GetAdminDashboard(c *gin.Context)
	GetHrDashboard(c *gin.Context)
	GetFinanceDashboard(c *gin.Context)
	GetHeatmap(c *gin.Context)
	GetDailyPulse(c *gin.Context)
}

type dashboardHandler struct {
	service service.DashboardService
}

func NewDashboardHandler(service service.DashboardService) DashboardHandler {
	return &dashboardHandler{service: service}
}

// @Summary Get Heatmap Data
// @Description Get dynamic heatmap data filtered by user and type
// @Tags Dashboard
// @Produce json
// @Param type query string true "Aktivitas (clockin, clockout, leave)"
// @Param user_id query int false "ID User (opsional)"
// @Param date_from query string false "Format: YYYY-MM-DD"
// @Param date_to query string false "Format: YYYY-MM-DD"
// @Security BearerAuth
// @Security CookieAuth
// @Router /api/v1/dashboards/heatmap [get]
func (h *dashboardHandler) GetHeatmap(c *gin.Context) {
	var query modelDto.HeatmapQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Filter tidak valid", 400, "error", err.Error()))
		return
	}

	// Default activity type
	if query.Type == "" {
		query.Type = "clockin"
	}

	tenantID := c.MustGet("tenant_id").(uint)
	data, err := h.service.GetHeatmapData(c.Request.Context(), tenantID, query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Gagal mengambil data heatmap", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Success", 200, "success", data))
}

// @Summary Get Admin Dashboard Data
// @Tags Dashboard
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Router /api/v1/dashboards/admin [get]
func (h *dashboardHandler) GetAdminDashboard(c *gin.Context) {
	currentUserID := c.MustGet("user_id").(uint)
	data, err := h.service.GetAdminDashboard(c.Request.Context(), currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch admin dashboard", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Success", 200, "success", data))
}

// @Summary Get HR Dashboard Data
// @Tags Dashboard
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Router /api/v1/dashboards/hr [get]
func (h *dashboardHandler) GetHrDashboard(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)
	currentUserID := c.MustGet("user_id").(uint)
	data, err := h.service.GetHrDashboard(c.Request.Context(), tenantID, currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch HR dashboard", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Success", 200, "success", data))
}

// @Summary Get Finance Dashboard Data
// @Tags Dashboard
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Router /api/v1/dashboards/finance [get]
func (h *dashboardHandler) GetFinanceDashboard(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)
	currentUserID := c.MustGet("user_id").(uint)
	data, err := h.service.GetFinanceDashboard(c.Request.Context(), tenantID, currentUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch finance dashboard", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Success", 200, "success", data))
}

// @Summary Get Daily Pulse Dashboard Data
// @Description Get real-time daily organization pulse for Manager Home
// @Tags Dashboard
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Router /api/v1/dashboards/hr/daily-pulse [get]
func (h *dashboardHandler) GetDailyPulse(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)
	data, err := h.service.GetDailyPulse(c.Request.Context(), tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch daily pulse", 500, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Success", 200, "success", data))
}
