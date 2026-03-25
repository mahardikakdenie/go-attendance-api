package handler

import (
	"net/http"

	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"

	"github.com/gin-gonic/gin"
)

type AttendanceHandler interface {
	CheckIn(c *gin.Context)
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

// @Summary Check-in Absensi
// @Description Endpoint untuk mencatat kehadiran karyawan
// @Tags Attendance
// @Accept json
// @Produce json
// @Param request body model.AttendanceRequest true "Data Karyawan ID"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/attendance/checkin [post]
func (h *attendanceHandler) CheckIn(c *gin.Context) {
	var req model.AttendanceRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Format data salah: " + err.Error()})
		return
	}

	response, err := h.service.CheckIn(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Berhasil memproses absensi",
		"data":    response,
	})
}

// @Summary Hello Test
// @Description Endpoint untuk mengecek status API
// @Tags Health
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/ping [get]
func (h *attendanceHandler) HelloTest(c *gin.Context) {
	response := gin.H{
		"data": "Halo dari Controller (Handler) yang sudah rapi!",
	}
	c.JSON(http.StatusOK, response)
}
