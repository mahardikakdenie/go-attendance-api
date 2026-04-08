package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type AttendanceHandler interface {
	RecordAttendance(c *gin.Context)
	GetAllAttendance(c *gin.Context)
	GetAttendanceSummary(c *gin.Context)
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

// @Summary Get Attendance Summary
// @Description Get summary of today's attendance with comparison
// @Tags Attendance
// @Produce json
// @Security CookieAuth
// @Success 200 {object} modelDto.BaseResponse{data=model.AttendanceSummaryResponse}
// @Failure 500 {object} modelDto.BaseResponse
// @Router /api/v1/attendance/summary [get]
func (h *attendanceHandler) GetAttendanceSummary(c *gin.Context) {
	ctx := c.Request.Context()

	tenantIDVal, _ := c.Get("tenant_id")
	var tenantID uint
	if tenantIDVal != nil {
		tenantID = tenantIDVal.(uint)
	}

	summary, err := h.service.GetSummary(ctx, tenantID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch summary", http.StatusInternalServerError, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Attendance summary fetched successfully", http.StatusOK, "success", summary))
}

// @Summary Record Attendance
// @Description Record clock-in / clock-out with face & location
// @Tags Attendance
// @Accept json
// @Produce json
// @Param request body model.AttendanceRequest true "Attendance Data"
// @Security CookieAuth
// @Success 200 {object} modelDto.BaseResponse
// @Failure 400 {object} modelDto.BaseResponse
// @Failure 401 {object} modelDto.BaseResponse
// @Router /api/v1/attendance [post]
func (h *attendanceHandler) RecordAttendance(c *gin.Context) {
	var req model.AttendanceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", http.StatusBadRequest, "error", err.Error()))
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.BuildErrorResponse("Unauthorized", http.StatusUnauthorized, "error", nil))
		return
	}

	userID := userIDVal.(uint)

	res, err := h.service.RecordAttendance(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Failed", http.StatusBadRequest, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Success", http.StatusOK, "success", res))
}

// @Summary Get All Attendance
// @Description Get attendance list with filter & pagination
// @Tags Attendance
// @Produce json
// @Param user_id query int false "User ID"
// @Param status query string false "Status"
// @Param date_from query string false "Start date (YYYY-MM-DD)"
// @Param date_to query string false "End date (YYYY-MM-DD)"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Param include query string false "Relations: user,tenant,setting"
// @Security CookieAuth
// @Success 200 {object} modelDto.BaseResponse{data=modelDto.AttendanceListResponse}
// @Failure 500 {object} modelDto.BaseResponse
// @Router /api/v1/attendance [get]
func (h *attendanceHandler) GetAllAttendance(c *gin.Context) {
	var filter model.AttendanceFilter

	ctx := c.Request.Context()

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

	var includes []string
	if inc := c.Query("include"); inc != "" {
		includes = strings.Split(inc, ",")
	}

	data, total, err := h.service.GetAllData(ctx, filter, includes, limit, offset)
	if err != nil {
		response := utils.BuildErrorResponse("Failed to fetch attendance data", http.StatusInternalServerError, "error", err.Error())
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	meta := modelDto.Meta{
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	response := utils.BuildResponse(
		"Attendance data fetched successfully",
		http.StatusOK,
		"success",
		modelDto.AttendanceListResponse{
			Data: data,
			Meta: meta,
		},
	)

	c.JSON(http.StatusOK, response)
}

// @Summary Health Check
// @Description Check API status
// @Tags Health
// @Produce json
// @Router /api/v1/ping [get]
func (h *attendanceHandler) HelloTest(c *gin.Context) {
	response := utils.BuildResponse("Health check success", http.StatusOK, "success", "API is running 🚀")
	c.JSON(http.StatusOK, response)
}
