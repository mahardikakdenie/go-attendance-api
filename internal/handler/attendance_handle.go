package handler

import (
	"net/http"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type AttendanceHandler interface {
	RecordAttendance(c *gin.Context)
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
