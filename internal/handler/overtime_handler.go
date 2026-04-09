package handler

import (
	"net/http"
	"strconv"
	"time"

	modelDto "go-attendance-api/internal/dto"
	"go-attendance-api/internal/model"
	"go-attendance-api/internal/service"
	"go-attendance-api/internal/utils"

	"github.com/gin-gonic/gin"
)

type OvertimeHandler interface {
	CreateRequest(c *gin.Context)
	ApproveRequest(c *gin.Context)
	RejectRequest(c *gin.Context)
	GetAll(c *gin.Context)
	GetByID(c *gin.Context)
}

type overtimeHandler struct {
	service service.OvertimeService
}

func NewOvertimeHandler(service service.OvertimeService) OvertimeHandler {
	return &overtimeHandler{
		service: service,
	}
}

// @Summary Create Overtime Request
// @Description Employee request overtime
// @Tags Overtime
// @Accept json
// @Produce json
// @Param request body model.CreateOvertimeRequest true "Overtime Data"
// @Security CookieAuth
// @Success 201 {object} modelDto.BaseResponse{data=model.OvertimeResponse}
// @Failure 400 {object} modelDto.BaseResponse
// @Router /api/v1/overtime [post]
func (h *overtimeHandler) CreateRequest(c *gin.Context) {
	var req model.CreateOvertimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request", http.StatusBadRequest, "error", err.Error()))
		return
	}

	userID := c.MustGet("user_id").(uint)
	tenantID := c.MustGet("tenant_id").(uint)

	res, err := h.service.CreateRequest(c.Request.Context(), userID, tenantID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Failed", http.StatusBadRequest, "error", err.Error()))
		return
	}

	c.JSON(http.StatusCreated, utils.BuildResponse("Overtime request created", http.StatusCreated, "success", res))
}

// @Summary Approve Overtime Request
// @Description Manager/Admin approve overtime
// @Tags Overtime
// @Accept json
// @Produce json
// @Param id path int true "Overtime ID"
// @Param request body model.ApproveOvertimeRequest true "Notes"
// @Security CookieAuth
// @Success 200 {object} modelDto.BaseResponse{data=model.OvertimeResponse}
// @Failure 400 {object} modelDto.BaseResponse
// @Router /api/v1/overtime/approve/{id} [post]
func (h *overtimeHandler) ApproveRequest(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req model.ApproveOvertimeRequest
	c.ShouldBindJSON(&req)

	adminID := c.MustGet("user_id").(uint)

	res, err := h.service.ApproveRequest(c.Request.Context(), uint(id), adminID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Failed", http.StatusBadRequest, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Overtime approved", http.StatusOK, "success", res))
}

// @Summary Reject Overtime Request
// @Description Manager/Admin reject overtime
// @Tags Overtime
// @Accept json
// @Produce json
// @Param id path int true "Overtime ID"
// @Param request body model.ApproveOvertimeRequest true "Notes"
// @Security CookieAuth
// @Success 200 {object} modelDto.BaseResponse{data=model.OvertimeResponse}
// @Failure 400 {object} modelDto.BaseResponse
// @Router /api/v1/overtime/reject/{id} [post]
func (h *overtimeHandler) RejectRequest(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req model.ApproveOvertimeRequest
	c.ShouldBindJSON(&req)

	adminID := c.MustGet("user_id").(uint)

	res, err := h.service.RejectRequest(c.Request.Context(), uint(id), adminID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Failed", http.StatusBadRequest, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Overtime rejected", http.StatusOK, "success", res))
}

// @Summary Get All Overtime
// @Description Get overtime list with filters
// @Tags Overtime
// @Produce json
// @Param status query string false "Status"
// @Param user_id query int false "User ID"
// @Param date_from query string false "YYYY-MM-DD"
// @Param date_to query string false "YYYY-MM-DD"
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Security CookieAuth
// @Success 200 {object} modelDto.BaseResponse{data=modelDto.AttendanceListResponse}
// @Failure 500 {object} modelDto.BaseResponse
// @Router /api/v1/overtime [get]
func (h *overtimeHandler) GetAll(c *gin.Context) {
	var filter model.OvertimeFilter
	filter.TenantID = c.MustGet("tenant_id").(uint)

	if status := c.Query("status"); status != "" {
		filter.Status = model.OvertimeStatus(status)
	}
	if uid := c.Query("user_id"); uid != "" {
		if val, err := strconv.Atoi(uid); err == nil {
			filter.UserID = uint(val)
		}
	}
	if df := c.Query("date_from"); df != "" {
		if t, err := time.Parse("2006-01-02", df); err == nil {
			filter.DateFrom = &t
		}
	}
	if dt := c.Query("date_to"); dt != "" {
		if t, err := time.Parse("2006-01-02", dt); err == nil {
			filter.DateTo = &t
		}
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	filter.Limit = limit
	filter.Offset = offset

	data, total, err := h.service.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed", http.StatusInternalServerError, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Success", http.StatusOK, "success", modelDto.AttendanceListResponse{
		Data: data,
		Meta: modelDto.Meta{Total: total, Limit: limit, Offset: offset},
	}))
}

// @Summary Get Overtime By ID
// @Tags Overtime
// @Produce json
// @Param id path int true "ID"
// @Security CookieAuth
// @Success 200 {object} modelDto.BaseResponse{data=model.OvertimeResponse}
// @Failure 404 {object} modelDto.BaseResponse
// @Router /api/v1/overtime/{id} [get]
func (h *overtimeHandler) GetByID(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	res, err := h.service.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, utils.BuildErrorResponse("Not found", http.StatusNotFound, "error", err.Error()))
		return
	}
	c.JSON(http.StatusOK, utils.BuildResponse("Success", http.StatusOK, "success", res))
}
