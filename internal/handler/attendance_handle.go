package handler

import (
	"math"
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
	GetAttendanceHistory(c *gin.Context)
	GetAttendanceSummary(c *gin.Context)
	GetTodayAttendance(c *gin.Context)
	HelloTest(c *gin.Context)
}

// @Summary Get Today Attendance
// @Description Get today's attendance status for the logged-in user
// @Tags Attendance
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} modelDto.TodayAttendanceResponse
// @Failure 401 {object} utils.APIResponse
// @Router /api/v1/attendance/today [get]
func (h *attendanceHandler) GetTodayAttendance(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, utils.BuildErrorResponse("Unauthorized", http.StatusUnauthorized, "error", nil))
		return
	}

	userID := userIDVal.(uint)

	res, err := h.service.GetTodayAttendance(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch data", 500, "error", err.Error()))
		return
	}

	if res == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    nil,
		})
		return
	}

	// Transform to TodayAttendanceResponse
	status := "On Time"
	if res.Status == model.StatusLate {
		status = "Late"
	}

	duration := "0h 0m"
	clockOutStr := ""
	if res.ClockOutTime != nil {
		diff := res.ClockOutTime.Sub(res.ClockInTime)
		hours := int(diff.Hours())
		mins := int(diff.Minutes()) % 60
		duration = strconv.Itoa(hours) + "h " + strconv.Itoa(mins) + "m"
		clockOutStr = res.ClockOutTime.Format("03:04 PM")
	} else {
		// If still working, calculate duration from now
		diff := time.Since(res.ClockInTime)
		hours := int(diff.Hours())
		mins := int(diff.Minutes()) % 60
		duration = strconv.Itoa(hours) + "h " + strconv.Itoa(mins) + "m"
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": modelDto.TodayAttendanceResponse{
			ClockInTime:  res.ClockInTime.Format("03:04 PM"),
			ClockOutTime: clockOutStr,
			Status:       status,
			Duration:     duration,
			Date:         res.ClockInTime.Format("2006-01-02"),
		},
	})
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
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=model.AttendanceSummaryResponse}
// @Failure 500 {object} utils.APIResponse
// @Router /api/v1/attendance/summary [get]
func (h *attendanceHandler) GetAttendanceSummary(c *gin.Context) {
	ctx := c.Request.Context()

	var filter model.AttendanceFilter

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.Atoi(userIDStr); err == nil {
			filter.UserID = uint(userID)
		}
	}

	if status := c.Query("status"); status != "" {
		filter.Status = model.AttendanceStatus(status)
	}

	if search := c.Query("search"); search != "" {
		filter.Search = search
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

	tenantIDVal, _ := c.Get("tenant_id")
	var tenantID uint
	if tenantIDVal != nil {
		tenantID = tenantIDVal.(uint)
	}

	summary, err := h.service.GetSummary(ctx, tenantID, filter)
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
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=model.AttendanceResponse}
// @Failure 400 {object} utils.APIResponse
// @Failure 401 {object} utils.APIResponse
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
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=modelDto.AttendanceListResponse}
// @Failure 500 {object} utils.APIResponse
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

	if search := c.Query("search"); search != "" {
		filter.Search = search
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

	requesterID := c.MustGet("user_id").(uint)

	data, total, err := h.service.GetAllData(ctx, requesterID, filter, includes, limit, offset)
	if err != nil {
		response := utils.BuildErrorResponse("Failed to fetch attendance data", http.StatusInternalServerError, "error", err.Error())
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	pagination := utils.Pagination{
		Total:       total,
		PerPage:     limit,
		CurrentPage: (offset / limit) + 1,
		LastPage:    int(math.Ceil(float64(total) / float64(limit))),
	}
	if pagination.LastPage == 0 {
		pagination.LastPage = 1
	}

	response := utils.BuildResponseWithPagination(
		"Attendance data fetched successfully",
		http.StatusOK,
		"success",
		data,
		pagination,
	)

	c.JSON(http.StatusOK, response)
}

// @Summary Get Attendance History
// @Description Get attendance history with specific format for dashboard
// @Tags Attendance
// @Produce json
// @Param limit query int false "Limit"
// @Param page query int false "Page"
// @Param status query string false "Status filter"
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=[]modelDto.AttendanceHistoryItem}
// @Router /api/v1/attendance/history [get]
func (h *attendanceHandler) GetAttendanceHistory(c *gin.Context) {
	ctx := c.Request.Context()

	// 1. Get User Info
	userID := c.MustGet("user_id").(uint)
	role := c.MustGet("role").(string)

	// 2. Parse Query Params
	limit := 10
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil {
			limit = val
		}
	}

	page := 1
	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil {
			page = val
		}
	}

	status := c.Query("status")
	search := c.Query("search")

	// 3. Prepare Filter
	var filter model.AttendanceFilter
	if status != "" {
		filter.Status = model.AttendanceStatus(status)
	}

	if search != "" {
		filter.Search = search
	}

	// Logic: If not admin/hr, only show own records
	isAdmin := role == "superadmin" || role == "admin" || role == "hr"
	if !isAdmin {
		filter.UserID = userID
	}

	offset := (page - 1) * limit

	// 4. Fetch Data
	data, total, err := h.service.GetAllData(ctx, userID, filter, []string{"user"}, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch history", http.StatusInternalServerError, "error", err.Error()))
		return
	}

	// 5. Transform to History Items
	items := make([]modelDto.AttendanceHistoryItem, 0)
	for _, a := range data {
		item := modelDto.AttendanceHistoryItem{
			ID:       a.ID.String(),
			Date:     a.ClockInTime.Format("2006-01-02"),
			ClockIn:  a.ClockInTime.Format("03:04 PM"),
			Status:   string(a.Status),
			Location: "Main Office", // Default mock as requested
			Overtime: "0h 0m",       // Default mock as requested
		}

		if a.ClockOutTime != nil {
			item.ClockOut = a.ClockOutTime.Format("03:04 PM")

			// Simple overtime calculation: anything after 5:00 PM
			fivePM := time.Date(a.ClockOutTime.Year(), a.ClockOutTime.Month(), a.ClockOutTime.Day(), 17, 0, 0, 0, a.ClockOutTime.Location())
			if a.ClockOutTime.After(fivePM) {
				diff := a.ClockOutTime.Sub(fivePM)
				hours := int(diff.Hours())
				mins := int(diff.Minutes()) % 60
				item.Overtime = strconv.Itoa(hours) + "h " + strconv.Itoa(mins) + "m"
			}
		}

		if a.User != nil {
			item.Employee.ID = strconv.Itoa(int(a.User.ID))
			item.Employee.Name = a.User.Name
			item.Employee.Avatar = a.User.MediaUrl
		}

		// Map status to frontend friendly labels
		switch a.Status {
		case model.StatusWorking:
			item.Status = "On Time"
		case model.StatusDone:
			item.Status = "On Time"
		case model.StatusLate:
			item.Status = "Late"
		}

		items = append(items, item)
	}

	// 6. Build Response
	lastPage := int(math.Ceil(float64(total) / float64(limit)))
	if lastPage == 0 {
		lastPage = 1
	}

	pagination := utils.Pagination{
		Total:       total,
		PerPage:     limit,
		CurrentPage: page,
		LastPage:    lastPage,
	}

	c.JSON(http.StatusOK, utils.BuildResponseWithPagination(
		"Attendance history fetched successfully",
		http.StatusOK,
		"success",
		items,
		pagination,
	))
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
