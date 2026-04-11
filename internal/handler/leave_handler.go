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

type LeaveHandler interface {
	RequestLeave(c *gin.Context)
	GetLeaveHistory(c *gin.Context)
	GetLeaveBalances(c *gin.Context)
	ApproveLeave(c *gin.Context)
	RejectLeave(c *gin.Context)
}

type leaveHandler struct {
	service service.LeaveService
}

func NewLeaveHandler(service service.LeaveService) LeaveHandler {
	return &leaveHandler{service: service}
}

type ReviewLeaveRequest struct {
	Notes string `json:"notes"`
}

// @Summary Approve Leave
// @Description Approve a pending leave request (HR/Manager only)
// @Tags Leaves
// @Accept json
// @Produce json
// @Param id path int true "Leave ID"
// @Param body body ReviewLeaveRequest false "Notes"
// @Security BearerAuth
// @Security CookieAuth
// @Router /api/v1/leaves/approve/{id} [post]
func (h *leaveHandler) ApproveLeave(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req ReviewLeaveRequest
	_ = c.ShouldBindJSON(&req)

	approverID := c.MustGet("user_id").(uint)

	if err := h.service.ApproveLeave(c.Request.Context(), approverID, uint(id), req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse(err.Error(), 500, "error", nil))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Leave approved successfully", 200, "success", nil))
}

// @Summary Reject Leave
// @Description Reject a pending leave request (HR/Manager only)
// @Tags Leaves
// @Accept json
// @Produce json
// @Param id path int true "Leave ID"
// @Param body body ReviewLeaveRequest false "Notes"
// @Security BearerAuth
// @Security CookieAuth
// @Router /api/v1/leaves/reject/{id} [post]
func (h *leaveHandler) RejectLeave(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var req ReviewLeaveRequest
	_ = c.ShouldBindJSON(&req)

	approverID := c.MustGet("user_id").(uint)

	if err := h.service.RejectLeave(c.Request.Context(), approverID, uint(id), req.Notes); err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse(err.Error(), 500, "error", nil))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Leave rejected successfully", 200, "success", nil))
}

// @Summary Request Leave
// @Description Submit a new leave request
// @Tags Leaves
// @Accept json
// @Produce json
// @Param request body model.LeaveRequest true "Leave Request Data"
// @Security BearerAuth
// @Security CookieAuth
// @Success 201 {object} utils.APIResponse{data=model.LeaveResponse}
// @Failure 400 {object} utils.APIResponse
// @Router /api/v1/leaves/request [post]
func (h *leaveHandler) RequestLeave(c *gin.Context) {
	var req model.LeaveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse("Invalid request body", 400, "error", err.Error()))
		return
	}

	userID := c.MustGet("user_id").(uint)
	tenantID := c.MustGet("tenant_id").(uint)

	res, err := h.service.RequestLeave(c.Request.Context(), userID, tenantID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.BuildErrorResponse(err.Error(), 400, "error", nil))
		return
	}

	c.JSON(http.StatusCreated, utils.BuildResponse("Leave request submitted successfully", 201, "success", res))
}

// @Summary Get Leave History
// @Description Get leave history for the logged-in user
// @Tags Leaves
// @Produce json
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=[]model.LeaveResponse}
// @Router /api/v1/leaves [get]
func (h *leaveHandler) GetLeaveHistory(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	var filter model.LeaveFilter
	// Logic: If not admin/hr, only show own records (scoping will handle the rest)
	role := c.MustGet("role").(string)
	isAdmin := role == "superadmin" || role == "admin" || role == "hr"
	if !isAdmin {
		filter.UserID = userID
	}

	res, total, err := h.service.GetLeaveHistory(c.Request.Context(), userID, filter, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch history", 500, "error", err.Error()))
		return
	}

	if limit <= 0 {
		limit = 10
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

	c.JSON(http.StatusOK, utils.BuildResponseWithPagination("History fetched successfully", 200, "success", res, pagination))
}

// @Summary Get Leave Balances
// @Description Get current year leave balances for the logged-in user
// @Tags Leaves
// @Produce json
// @Security BearerAuth
// @Security CookieAuth
// @Success 200 {object} utils.APIResponse{data=[]model.LeaveBalance}
// @Router /api/v1/leaves/balances [get]
func (h *leaveHandler) GetLeaveBalances(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	res, err := h.service.GetLeaveBalances(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.BuildErrorResponse("Failed to fetch balances", 500, "error", err.Error()))
		return
	}

	c.JSON(http.StatusOK, utils.BuildResponse("Balances fetched successfully", 200, "success", res))
}
