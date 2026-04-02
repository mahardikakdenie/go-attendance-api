package handler

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type AttendanceHandler interface {
	RecordAttendance(c *gin.Context)
	GetAllAttendance(c *gin.Context)
	HelloTest(c *gin.Context)
}

type attendanceHandler struct {
	service service.AttendanceService
}

func NewAttendanceHandler(service service.AttendanceService) AttendanceHandler {
	return &attendanceHandler{
		service: service,
	}
}

// @Summary Record Attendance
// @Description Record clock-in / clock-out with face & location
// @Tags Attendance
// @Accept json
// @Produce json
// @Param request body model.AttendanceRequest true "Attendance Data"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/attendance [post]
func (h *attendanceHandler) RecordAttendance(c *gin.Context) {
	var req model.AttendanceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, utils.BuildErrorResponse("Invalid request", 400, "error", err.Error()))
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(401, utils.BuildErrorResponse("Unauthorized", 401, "error", nil))
		return
	}

	userID := userIDVal.(uint)

	res, err := h.service.RecordAttendance(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(400, utils.BuildErrorResponse("Failed", 400, "error", err.Error()))
		return
	}

	c.JSON(200, utils.BuildResponse("Success", 200, "success", res))
}

// @Summary Get All Attendance
// @Description Get attendance list with filter & pagination
// @Tags Attendance
// @Produce json
// @Param user_id query int false "User ID"
// @Param status query string false "Status (working, done, late)"
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Param limit query int false "Limit (default 10)"
// @Param offset query int false "Offset (default 0)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/attendance [get]
func (h *attendanceHandler) GetAllAttendance(c *gin.Context) {
	var filter model.AttendanceFilter

	ctx := context.Background()

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.Atoi(userIDStr); err == nil {
			filter.UserID = uint(userID)
		}
	}

	if status := c.Query("status"); status != "" {
		filter.Status = model.AttendanceStatus(status)
	}

	if dateFrom := c.Query("date_from"); dateFrom != "" {
		if t, err := time.Parse("2006-01-02", dateFrom); err == nil {
			filter.DateFrom = &t
		}
	}

	if dateTo := c.Query("date_to"); dateTo != "" {
		if t, err := time.Parse("2006-01-02", dateTo); err == nil {
			filter.DateTo = &t
		}
	}

	limit := 10
	offset := 0

	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	if o := c.Query("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil {
			offset = val
		}
	}

	data, total, err := h.service.GetAllData(ctx, filter, limit, offset)
	if err != nil {
		response := utils.BuildErrorResponse("Failed to fetch attendance data", http.StatusInternalServerError, "error", err.Error())
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	meta := map[string]interface{}{
		"total":  total,
		"limit":  limit,
		"offset": offset,
	}

	response := utils.BuildResponse("Attendance data fetched successfully", http.StatusOK, "success", gin.H{
		"data": data,
		"meta": meta,
	})

	c.JSON(http.StatusOK, response)
}

// @Summary Health Check
// @Description Check API status
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/ping [get]
func (h *attendanceHandler) HelloTest(c *gin.Context) {
	response := utils.BuildResponse("Health check success", http.StatusOK, "success", "API is running 🚀")
	c.JSON(http.StatusOK, response)
}
