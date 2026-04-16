package handler

import (
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"
	"math"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type AttendanceCorrectionHandler interface {
	RequestCorrection(c *gin.Context)
	GetCorrections(c *gin.Context)
	ApproveCorrection(c *gin.Context)
	RejectCorrection(c *gin.Context)
}

type attendanceCorrectionHandler struct {
	service service.AttendanceCorrectionService
}

func NewAttendanceCorrectionHandler(service service.AttendanceCorrectionService) AttendanceCorrectionHandler {
	return &attendanceCorrectionHandler{service: service}
}

// @Summary Request Attendance Correction
// @Description Submit a request to correct missed attendance
// @Tags AttendanceCorrections
// @Accept json
// @Produce json
// @Param body body model.CreateCorrectionRequest true "Request Body"
// @Security BearerAuth
// @Router /api/v1/attendance/corrections [post]
func (h *attendanceCorrectionHandler) RequestCorrection(c *gin.Context) {
	var req model.CreateCorrectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request body", 400, "error", err.Error()))
		return
	}

	userID := c.MustGet("user_id").(uint)
	tenantID := c.MustGet("tenant_id").(uint)

	res, err := h.service.RequestCorrection(c.Request.Context(), userID, tenantID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse(err.Error(), 400, "error", nil))
		return
	}

	c.JSON(http.StatusCreated, utils.BuildResponse("Correction request submitted", 201, "success", res))
}

// @Summary Get Attendance Corrections
// @Description Get list of attendance correction requests
// @Tags AttendanceCorrections
// @Produce json
// @Param status query string false "Filter by status (PENDING, APPROVED, REJECTED)"
// @Param page query int false "Page number (default: 1)"
// @Param limit query int false "Items per page (default: 10)"
// @Security BearerAuth
// @Router /api/v1/attendance/corrections [get]
func (h *attendanceCorrectionHandler) GetCorrections(c *gin.Context) {
	tenantID := c.MustGet("tenant_id").(uint)
	userID := uint(0)
	
	// If not admin/hr, only show own requests
	role := c.MustGet("role").(string)
	if role != "admin" && role != "hr" && role != "superadmin" {
		userID = c.MustGet("user_id").(uint)
	}

	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	offset := (page - 1) * limit

	res, total, err := h.service.GetCorrections(c.Request.Context(), tenantID, userID, status, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch requests", 500, "error", err.Error()))
		return
	}

	pagination := utils.Pagination{
		Total:       total,
		PerPage:     limit,
		CurrentPage: page,
		LastPage:    int(math.Ceil(float64(total) / float64(limit))),
	}
	if pagination.LastPage == 0 {
		pagination.LastPage = 1
	}

	c.JSON(http.StatusOK, utils.BuildResponseWithPagination("Requests fetched successfully", 200, "success", res, pagination))
}

// @Summary Approve Attendance Correction
// @Description Approve a pending attendance correction request
// @Tags AttendanceCorrections
// @Accept json
// @Produce json
// @Param id path int true "Request ID"
// @Param body body model.ReviewCorrectionRequest true "Admin Notes"
// @Security BearerAuth
// @Router /api/v1/attendance/corrections/{id}/approve [post]
func (h *attendanceCorrectionHandler) ApproveCorrection(c *gin.Context) {
	idParam := c.Param("id")
	id, _ := strconv.Atoi(idParam)
	adminID := c.MustGet("user_id").(uint)

	var req model.ReviewCorrectionRequest
	_ = c.ShouldBindJSON(&req)

	err := h.service.ApproveCorrection(c.Request.Context(), uint(id), adminID, req.AdminNotes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse(err.Error(), 500, "error", nil))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Request approved successfully", 200, "success", nil))
}

// @Summary Reject Attendance Correction
// @Description Reject a pending attendance correction request
// @Tags AttendanceCorrections
// @Accept json
// @Produce json
// @Param id path int true "Request ID"
// @Param body body model.ReviewCorrectionRequest true "Admin Notes"
// @Security BearerAuth
// @Router /api/v1/attendance/corrections/{id}/reject [post]
func (h *attendanceCorrectionHandler) RejectCorrection(c *gin.Context) {
	idParam := c.Param("id")
	id, _ := strconv.Atoi(idParam)
	adminID := c.MustGet("user_id").(uint)

	var req model.ReviewCorrectionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Notes required for rejection", 400, "error", nil))
		return
	}

	err := h.service.RejectCorrection(c.Request.Context(), uint(id), adminID, req.AdminNotes)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse(err.Error(), 500, "error", nil))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Request rejected successfully", 200, "success", nil))
}
