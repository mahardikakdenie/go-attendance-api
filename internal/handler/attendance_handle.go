package handler

import (
	"net/http"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"

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
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/attendance [post]
func (h *attendanceHandler) RecordAttendance(c *gin.Context) {
	var req model.AttendanceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format: " + err.Error()})
		return
	}

	response, err := h.service.RecordAttendance(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Attendance recorded successfully",
		"data":    response,
	})
}

// @Summary Health Check
// @Description Endpoint to check API status
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/ping [get]
func (h *attendanceHandler) HelloTest(c *gin.Context) {
	response := gin.H{
		"data": "Hello from the clean architecture Handler!",
	}
	c.JSON(http.StatusOK, response)
}
