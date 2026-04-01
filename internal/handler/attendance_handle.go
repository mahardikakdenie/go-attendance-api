package handler

import (
	"net/http"
	"strconv"

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
// @Description Endpoint to record employee clock-in and clock-out with geolocation
// @Tags Attendance
// @Accept json
// @Produce json
// @Param request body model.AttendanceRequest true "Attendance Data (Action: clock_in/clock_out)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/attendance [post]
func (h *attendanceHandler) RecordAttendance(c *gin.Context) {
	var req model.AttendanceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response := utils.BuildErrorResponse("Invalid request format", http.StatusBadRequest, "error", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	res, err := h.service.RecordAttendance(req)
	if err != nil {
		response := utils.BuildErrorResponse("Failed to record attendance", http.StatusBadRequest, "error", err.Error())
		c.JSON(http.StatusBadRequest, response)
		return
	}

	response := utils.BuildResponse("Attendance recorded successfully", http.StatusOK, "success", res)
	c.JSON(http.StatusOK, response)
}

// @Summary Get All Attendance Data
// @Description Endpoint to retrieve all attendance records
// @Tags Attendance
// @Produce json
// @Param user_id query int false "Filter by User ID"
// @Param status query string false "Filter by Status (e.g., On Time, Late)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/attendance [get]
func (h *attendanceHandler) GetAllAttendance(c *gin.Context) {
	var filter model.AttendanceFilter

	userIDStr := c.Query("user_id")
	if userIDStr != "" {
		userID, err := strconv.Atoi(userIDStr)
		if err == nil {
			filter.UserID = userID
		}
	}

	filter.Status = c.Query("status")

	data, err := h.service.GetAllData(filter)
	if err != nil {
		response := utils.BuildErrorResponse("Failed to fetch attendance data", http.StatusInternalServerError, "error", err.Error())
		c.JSON(http.StatusInternalServerError, response)
		return
	}

	response := utils.BuildResponse("Attendance data fetched successfully", http.StatusOK, "success", data)
	c.JSON(http.StatusOK, response)
}

// @Summary Health Check
// @Description Endpoint to check API status
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/ping [get]
func (h *attendanceHandler) HelloTest(c *gin.Context) {
	response := utils.BuildResponse("Health check success", http.StatusOK, "success", "Hello from the clean architecture Handler!")
	c.JSON(http.StatusOK, response)
}
